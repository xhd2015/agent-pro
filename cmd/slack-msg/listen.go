package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
)

type listenConfig struct {
	BotToken              string
	AppToken              string
	ConfigPath            string
	ConfigDisplay         string
	SlackConfig           *slackutil.SlackConfig
	Channels              []string
	RequireMention        bool
	AllowFrom             []string
	SessionMode           string
	IdleTimeout           time.Duration
	AgentRunner           string
	AgentRunnerConfigHome string
	ReplyPrefix           string
	LockFile              string
}

type listener struct {
	api         *slack.Client
	filter      filterConfig
	sessions    *sessionStore
	agent       agentOptions
	replyPrefix string
	sessionMode string
	processMu   sync.Mutex
}

func (l *listener) handleEventsAPI(evt *socketmode.Event, smClient *socketmode.Client) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return
	}
	if evt.Request != nil {
		_ = smClient.Ack(*evt.Request)
	}
	if eventsAPIEvent.Type != slackevents.CallbackEvent {
		return
	}

	switch ev := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		l.dispatchInbound(inboundMessage{
			EventType: "app_mention",
			ChannelID: ev.Channel,
			UserID:    ev.User,
			Text:      ev.Text,
			TS:        ev.TimeStamp,
			ThreadTS:  ev.ThreadTimeStamp,
		})
	case *slackevents.MessageEvent:
		if ev.SubType != "" {
			return
		}
		l.dispatchInbound(inboundMessage{
			EventType: "message",
			ChannelID: ev.Channel,
			UserID:    ev.User,
			Text:      ev.Text,
			TS:        ev.TimeStamp,
			ThreadTS:  ev.ThreadTimeStamp,
		})
	}
}

func (l *listener) dispatchInbound(msg inboundMessage) {
	if !shouldProcess(msg, l.filter) {
		return
	}

	l.processMu.Lock()
	defer l.processMu.Unlock()

	threadTS := rootThreadTS(msg)
	sid := sessionID(msg.ChannelID, threadTS)

	var reply string
	var err error

	switch l.sessionMode {
	case "stateless":
		reply, err = runAgent(msg.Text, agentOptions{
			Runner:           l.agent.Runner,
			RunnerConfigHome: l.agent.RunnerConfigHome,
			SessionMode:      "stateless",
		})
	default:
		sess, exists := l.sessions.lookup(msg.ChannelID, threadTS)
		if exists && sess.Started {
			reply, err = runAgent(msg.Text, agentOptions{
				SessionMode: "thread",
				SessionID:   sid,
				IsFollowUp:  true,
			})
			l.sessions.touch(msg.ChannelID, threadTS, sid, true)
		} else {
			reply, err = runAgent(msg.Text, agentOptions{
				Runner:           l.agent.Runner,
				RunnerConfigHome: l.agent.RunnerConfigHome,
				SessionMode:      "thread",
				SessionID:        sid,
			})
			l.sessions.touch(msg.ChannelID, threadTS, sid, true)
		}
	}
	if err != nil {
		return
	}

	text := l.replyPrefix + reply
	_, _, postErr := l.api.PostMessage(
		msg.ChannelID,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTS),
	)
	if postErr != nil {
		return
	}
}

func resolveBotUserID(api *slack.Client, auth *slack.AuthTestResponse) (string, error) {
	if auth == nil {
		return "", fmt.Errorf("auth response is nil")
	}

	// Real bot tokens: auth.test user_id is the bot user and users.info reports is_bot.
	if auth.UserID != "" {
		user, err := api.GetUserInfo(auth.UserID)
		if err == nil && user != nil && user.IsBot {
			return auth.UserID, nil
		}
	}

	// slacktest auth.test returns a human user_id; bots.info still exposes the bot id.
	botParam := auth.BotID
	if botParam == "" {
		botParam = auth.UserID
	}
	if botParam != "" {
		bot, err := api.GetBotInfo(slack.GetBotInfoParameters{Bot: botParam})
		if err == nil && bot != nil {
			if bot.UserID != "" {
				return bot.UserID, nil
			}
			if bot.ID != "" {
				return bot.ID, nil
			}
		}
	}

	if auth.UserID != "" {
		return auth.UserID, nil
	}
	return "", fmt.Errorf("bot user id not found")
}

func resolveChannelFilter(api *slack.Client, cfg *slackutil.SlackConfig, channels []string) ([]string, error) {
	if len(channels) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(channels))
	for _, ch := range channels {
		resolved, err := slackutil.ResolveChannel(api, cfg, ch)
		if err != nil {
			return nil, err
		}
		out = append(out, resolved)
	}
	return out, nil
}

func runListen(cfg listenConfig) error {
	lock, err := maybeAcquireLock(cfg.LockFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if lock != nil {
		defer lock.release()
	}

	fmt.Fprintf(os.Stderr, "Using config from: %s\n", cfg.ConfigDisplay)

	apiOpts := []slack.Option{slack.OptionAppLevelToken(cfg.AppToken)}
	api := slackutil.NewAPIClient(cfg.BotToken, apiOpts...)

	auth, err := api.AuthTest()
	if err != nil {
		return fmt.Errorf("auth test failed: %w", err)
	}
	botUserID, err := resolveBotUserID(api, auth)
	if err != nil {
		return fmt.Errorf("resolve bot user id: %w", err)
	}

	channelFilter, err := resolveChannelFilter(api, cfg.SlackConfig, cfg.Channels)
	if err != nil {
		return err
	}

	allowFrom := cfg.AllowFrom
	if len(allowFrom) == 0 {
		allowFrom = []string{"*"}
	}

	client := socketmode.New(api)
	l := &listener{
		api: api,
		filter: filterConfig{
			BotUserID:      botUserID,
			RequireMention: cfg.RequireMention,
			AllowFrom:      allowFrom,
			ChannelFilter:  channelFilter,
		},
		sessions: newSessionStore(cfg.IdleTimeout),
		agent: agentOptions{
			Runner:           cfg.AgentRunner,
			RunnerConfigHome: cfg.AgentRunnerConfigHome,
		},
		replyPrefix: cfg.ReplyPrefix,
		sessionMode: cfg.SessionMode,
	}

	done := make(chan struct{})
	defer close(done)
	go l.sessions.pruneLoop(done)

	handler := socketmode.NewSocketmodeHandler(client)
	handler.Handle(socketmode.EventTypeEventsAPI, l.handleEventsAPI)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	err = handler.RunEventLoopContext(ctx)
	if err != nil && errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func maybeAcquireLock(path string) (*processLock, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	return acquireLock(path)
}
