package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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
	dataRoot    string
	// configPathAbs is the absolute path listen resolved for --config (may be empty).
	configPathAbs string
	processMu     sync.Mutex
	dedupeMu      sync.Mutex
	dedupeSeen    map[string]struct{}
	userCacheMu   sync.Mutex
	userCache     map[string]string
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

	// Dedupe app_mention + message dual delivery on the same channel+ts.
	if !l.claimEvent(msg.ChannelID, msg.TS) {
		return
	}

	l.processMu.Lock()
	defer l.processMu.Unlock()

	threadTS := rootThreadTS(msg)
	sid := conversationSessionID(msg.ChannelID, msg.UserID)
	cleaned := stripBotMention(msg.Text, l.filter.BotUserID)
	display := l.userDisplayName(msg.UserID)

	l.logAccepted(msg, display, cleaned)

	switch l.sessionMode {
	case "stateless":
		prompt := cleaned
		if prompt == "" {
			prompt = strings.TrimSpace(msg.Text)
		}
		if prompt == "" {
			prompt = "(empty message)"
		}
		fmt.Fprintf(os.Stderr, "agent run start session-mode=stateless\n")
		reply, err := runAgentStateless(prompt, l.agent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "agent run failed: %v\n", err)
			return
		}
		text := l.replyPrefix + reply
		_, _, _ = l.api.PostMessage(
			msg.ChannelID,
			slack.MsgOptionText(text, false),
			slack.MsgOptionTS(threadTS),
		)
	default:
		// Thread mode: upsert durable map, append inbound log, SYSTEM.md inject,
		// interactive open with SLACK_MSG_* env; no PostMessage of agent body.
		preview := cleaned
		if preview == "" {
			preview = strings.TrimSpace(msg.Text)
		}
		if err := upsertSessionEntry(l.dataRoot, durableSessionEntry{
			SessionID:          sid,
			ChannelID:          msg.ChannelID,
			ThreadTS:           threadTS,
			ConfigPath:         l.configPathAbs,
			Kind:               "channel",
			ReplyMode:          "channel",
			LastMessagePreview: truncateRunes(preview, 80),
		}); err != nil {
			fmt.Fprintf(os.Stderr, "upsert sessions.json failed: %v\n", err)
			return
		}
		if err := appendSessionMessage(l.dataRoot, sid, sessionLogEntry{
			MessageID: msg.TS,
			TS:        msg.TS,
			User:      msg.UserID,
			Text:      cleaned,
			Direction: "in",
		}); err != nil {
			fmt.Fprintf(os.Stderr, "append messages.jsonl failed: %v\n", err)
			return
		}

		sysPath := sessionSystemMDPath(l.dataRoot, sid)
		if err := writeSystemPrompt(sysPath, systemPromptContext{
			SessionID: sid,
			ChannelID: msg.ChannelID,
			ThreadTS:  threadTS,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "write SYSTEM.md failed: %v\n", err)
			return
		}
		absSys := sysPath
		if p, err := filepath.Abs(sysPath); err == nil {
			absSys = p
		}
		inject := formatOpenInject(openInjectContext{
			SessionID:        sid,
			ChannelID:        msg.ChannelID,
			ThreadTS:         threadTS,
			FromDisplay:      display,
			FromUserID:       msg.UserID,
			SystemPromptPath: absSys,
			UserMessage:      cleaned,
		})

		openOpts := l.agent
		openOpts.Env = append([]string(nil), openOpts.Env...)
		openOpts.Env = append(openOpts.Env, envSlackMsgSessionID+"="+sid)
		if l.configPathAbs != "" {
			openOpts.Env = append(openOpts.Env, envSlackMsgConfig+"="+l.configPathAbs)
		}
		if entry, lookupErr := lookupSessionEntry(l.dataRoot, sid); lookupErr == nil && entry != nil {
			if d := strings.TrimSpace(entry.Dir); d != "" {
				openOpts.WorkspaceDir = d
			}
		}

		fmt.Fprintf(os.Stderr, "agent open start session=%s\n", sid)
		if err := runAgentInteractiveOpen(inject, sid, openOpts); err != nil {
			fmt.Fprintf(os.Stderr, "agent open failed: %v\n", err)
			return
		}
		l.sessions.touch(msg.ChannelID, threadTS, sid, true)
	}
}

func (l *listener) claimEvent(channelID, ts string) bool {
	key := channelID + ":" + ts
	if channelID == "" || ts == "" {
		return true
	}
	l.dedupeMu.Lock()
	defer l.dedupeMu.Unlock()
	if l.dedupeSeen == nil {
		l.dedupeSeen = make(map[string]struct{})
	}
	if _, seen := l.dedupeSeen[key]; seen {
		return false
	}
	l.dedupeSeen[key] = struct{}{}
	return true
}

func (l *listener) userDisplayName(userID string) string {
	if userID == "" {
		return ""
	}
	l.userCacheMu.Lock()
	if l.userCache != nil {
		if name, ok := l.userCache[userID]; ok {
			l.userCacheMu.Unlock()
			return name
		}
	}
	l.userCacheMu.Unlock()

	name := userID
	if user, err := l.api.GetUserInfo(userID); err == nil && user != nil {
		switch {
		case user.Profile.DisplayName != "":
			name = user.Profile.DisplayName
		case user.Name != "":
			name = user.Name
		case user.RealName != "":
			name = user.RealName
		}
	}

	l.userCacheMu.Lock()
	if l.userCache == nil {
		l.userCache = make(map[string]string)
	}
	l.userCache[userID] = name
	l.userCacheMu.Unlock()
	return name
}

func (l *listener) logAccepted(msg inboundMessage, display, cleanedText string) {
	snippet := cleanedText
	if snippet == "" {
		snippet = strings.TrimSpace(msg.Text)
	}
	if display == "" {
		display = msg.UserID
	}
	// Operator log: kind, display, channel, ts, text excerpt.
	fmt.Fprintf(os.Stderr, "%s  %s  %s  ts=%s  %q\n",
		msg.EventType, display, msg.ChannelID, msg.TS, truncateRunes(snippet, 120))
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
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

// resolveBotIdentity returns display name and optional bot_id for the banner.
func resolveBotIdentity(api *slack.Client, auth *slack.AuthTestResponse, botUserID string) (name, botID string) {
	if auth != nil {
		botID = auth.BotID
		name = auth.User
	}
	// Prefer bots.info for TestSlackBot-style names under slacktest.
	botParam := botID
	if botParam == "" {
		botParam = botUserID
	}
	if botParam == "" && auth != nil {
		botParam = auth.UserID
	}
	if botParam != "" {
		if bot, err := api.GetBotInfo(slack.GetBotInfoParameters{Bot: botParam}); err == nil && bot != nil {
			if bot.Name != "" {
				name = bot.Name
			}
			if bot.ID != "" && botID == "" {
				botID = bot.ID
			}
		}
	}
	if name == "" && botUserID != "" {
		if user, err := api.GetUserInfo(botUserID); err == nil && user != nil {
			if user.Name != "" {
				name = user.Name
			} else if user.RealName != "" {
				name = user.RealName
			}
		}
	}
	return name, botID
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

func printStartupBanner(cfg listenConfig, auth *slack.AuthTestResponse, botUserID, botName, botID string) {
	// Banner goes to stdout (line-flushed) so daemon harness probes that snapshot
	// buffers before SIGTERM still observe identity + lock lines. Stderr remains
	// available for operator event logs after connect.
	out := os.Stdout
	fmt.Fprintf(out, "Using config from: %s\n", cfg.ConfigDisplay)

	team := ""
	teamID := ""
	if auth != nil {
		team = auth.Team
		teamID = auth.TeamID
	}
	if team != "" || teamID != "" {
		if team != "" && teamID != "" {
			fmt.Fprintf(out, "team: %s (%s)\n", team, teamID)
		} else if team != "" {
			fmt.Fprintf(out, "team: %s\n", team)
		} else {
			fmt.Fprintf(out, "team: %s\n", teamID)
		}
	}

	if botName != "" && botUserID != "" {
		if botID != "" {
			fmt.Fprintf(out, "bot: %s (%s) bot_id=%s\n", botName, botUserID, botID)
		} else {
			fmt.Fprintf(out, "bot: %s (%s)\n", botName, botUserID)
		}
	} else if botUserID != "" {
		if botID != "" {
			fmt.Fprintf(out, "bot: %s bot_id=%s\n", botUserID, botID)
		} else {
			fmt.Fprintf(out, "bot: %s\n", botUserID)
		}
	}

	sessionMode := cfg.SessionMode
	if sessionMode == "" {
		sessionMode = "thread"
	}
	fmt.Fprintf(out, "session-mode: %s\n", sessionMode)
	fmt.Fprintf(out, "require-mention: %v\n", cfg.RequireMention)
	agentRunner := cfg.AgentRunner
	if agentRunner == "" {
		agentRunner = "grok-tty"
	}
	fmt.Fprintf(out, "agent-runner: %s\n", agentRunner)

	lockDisplay := strings.TrimSpace(cfg.LockFile)
	if lockDisplay == "" {
		lockDisplay = "(none)"
	}
	fmt.Fprintf(out, "lock-file: %s\n", lockDisplay)
	_ = out.Sync()
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

	apiOpts := []slack.Option{slack.OptionAppLevelToken(cfg.AppToken)}
	api := slackutil.NewAPIClient(cfg.BotToken, apiOpts...)

	auth, err := api.AuthTest()
	if err != nil {
		// Config line on failure path so operators still see which config was used.
		fmt.Fprintf(os.Stderr, "Using config from: %s\n", cfg.ConfigDisplay)
		return fmt.Errorf("auth test failed: %w", err)
	}
	botUserID, err := resolveBotUserID(api, auth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Using config from: %s\n", cfg.ConfigDisplay)
		return fmt.Errorf("resolve bot user id: %w", err)
	}
	botName, botID := resolveBotIdentity(api, auth, botUserID)
	printStartupBanner(cfg, auth, botUserID, botName, botID)

	channelFilter, err := resolveChannelFilter(api, cfg.SlackConfig, cfg.Channels)
	if err != nil {
		return err
	}

	allowFrom := cfg.AllowFrom
	if len(allowFrom) == 0 {
		allowFrom = []string{"*"}
	}

	configPathAbs := resolveConfigPathAbs(cfg.ConfigPath)

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
		replyPrefix:   cfg.ReplyPrefix,
		sessionMode:   cfg.SessionMode,
		dataRoot:      defaultSlackLocalBotRoot(),
		configPathAbs: configPathAbs,
		dedupeSeen:    make(map[string]struct{}),
		userCache:     make(map[string]string),
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
