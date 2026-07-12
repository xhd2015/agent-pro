package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/slack-go/slack"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
	lessflags "github.com/xhd2015/less-flags"
)

const (
	envSlackMsgSessionID = "SLACK_MSG_SESSION_ID"
	envSlackMsgConfig    = "SLACK_MSG_CONFIG"
)

const sessionHelpText = `slack-msg session: session-bound reply and history.

Usage:
  slack-msg session <command> [options]

Commands:
  reply    Post a channel reply for the bound session
  history  Show local session message history

Options:
  -h, --help  Show help
`

const sessionReplyHelpText = `slack-msg session reply: post a channel-level reply for a bound Slack session.

Usage:
  slack-msg session reply [options] MESSAGE

Options:
  --session-id ID   Session id (env: SLACK_MSG_SESSION_ID)
  --config PATH     Config file (env: SLACK_MSG_CONFIG)
  --token TOKEN     Bot token override
  -h, --help        Show help
`

const sessionHistoryHelpText = `slack-msg session history: print local session message history.

Usage:
  slack-msg session history [options]

Options:
  --session-id ID     Session id (env: SLACK_MSG_SESSION_ID)
  --after-msg-id ID   Only messages after this id
  --limit N           Max messages
  --json              JSON output
  --config PATH       Config file (env: SLACK_MSG_CONFIG)
  -h, --help          Show help
`

func runSession(args []string) error {
	if len(args) == 0 {
		fmt.Print(sessionHelpText)
		return nil
	}
	switch args[0] {
	case "-h", "--help":
		fmt.Print(sessionHelpText)
		return nil
	case "reply":
		return runSessionReply(args[1:])
	case "history":
		return runSessionHistory(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("unknown option: %s", args[0])
		}
		return fmt.Errorf("unknown session command: %s", args[0])
	}
}

func runSessionReply(args []string) error {
	var (
		sessionIDFlag *string
		configFlag    *string
		tokenFlag     *string
	)

	remain, err := lessflags.String("--session-id", &sessionIDFlag).
		String("--config", &configFlag).
		String("--token", &tokenFlag).
		Help("-h,--help", sessionReplyHelpText).
		HelpNoExit().
		StopOnFirstArg().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	switch len(remain) {
	case 0:
		fmt.Fprintln(os.Stderr, "message required")
		os.Exit(1)
	case 1:
		// ok
	default:
		fmt.Fprintf(os.Stderr, "exactly one message required, got %d\n", len(remain))
		os.Exit(1)
	}
	message := remain[0]

	sessionID := strings.TrimSpace(flagString(sessionIDFlag))
	if sessionID == "" {
		sessionID = strings.TrimSpace(os.Getenv(envSlackMsgSessionID))
	}
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "session id required")
		os.Exit(1)
	}

	dataRoot := defaultSlackLocalBotRoot()
	entry, err := lookupSessionEntry(dataRoot, sessionID)
	if err != nil {
		if strings.Contains(err.Error(), "session not found") {
			fmt.Fprintln(os.Stderr, "session not found")
			os.Exit(1)
		}
		if strings.Contains(err.Error(), "session id required") {
			fmt.Fprintln(os.Stderr, "session id required")
			os.Exit(1)
		}
		return err
	}

	configPath := ""
	if configFlag != nil && *configFlag != "" {
		configPath = *configFlag
	} else if env := os.Getenv(envSlackMsgConfig); env != "" {
		configPath = env
	} else if entry.ConfigPath != "" {
		configPath = entry.ConfigPath
	}

	var cfg *slackutil.SlackConfig
	if configPath != "" {
		loaded, loadErr := slackutil.Load(configPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", loadErr)
			os.Exit(1)
		}
		cfg = loaded
	}

	token := slackutil.ResolveBotToken(flagString(tokenFlag), "SLACK_BOT_TOKEN", cfg)
	if token == "" {
		if cfg != nil && configPath != "" {
			fmt.Fprintf(os.Stderr, "botToken is empty in %s\n", slackutil.ConfigDisplayPath(configPath))
		} else {
			fmt.Fprintln(os.Stderr, "bot token required")
		}
		os.Exit(1)
	}

	channelID := strings.TrimSpace(entry.ChannelID)
	if channelID == "" {
		fmt.Fprintln(os.Stderr, "session has empty channel_id")
		os.Exit(1)
	}

	api := slackutil.NewAPIClient(token)
	// Channel top-level only: no MsgOptionTS / thread_ts.
	respChannel, ts, postErr := api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if postErr != nil {
		fmt.Fprintf(os.Stderr, "session reply failed: %v\n", postErr)
		os.Exit(1)
	}

	// Append outbound log (best-effort; do not fail reply on log error).
	_ = appendSessionMessage(dataRoot, sessionID, sessionLogEntry{
		MessageID: ts,
		TS:        ts,
		User:      "",
		Text:      message,
		Direction: "out",
	})
	// Refresh map preview.
	_ = upsertSessionEntry(dataRoot, durableSessionEntry{
		SessionID:          entry.SessionID,
		ChannelID:          entry.ChannelID,
		ThreadTS:           entry.ThreadTS,
		ConfigPath:         entry.ConfigPath,
		Kind:               entry.Kind,
		ReplyMode:          entry.ReplyMode,
		LastMessagePreview: truncateRunes(message, 80),
	})

	fmt.Printf("OK ts=%s channel=%s\n", ts, respChannel)
	return nil
}

func runSessionHistory(args []string) error {
	var (
		sessionIDFlag  *string
		afterMsgIDFlag *string
		limitFlag      *int
		configFlag     *string
		jsonFlag       bool
	)

	remain, err := lessflags.String("--session-id", &sessionIDFlag).
		String("--after-msg-id", &afterMsgIDFlag).
		Int("--limit", &limitFlag).
		String("--config", &configFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", sessionHistoryHelpText).
		HelpNoExit().
		Parse(args)
	if errors.Is(err, lessflags.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remain) > 0 {
		return fmt.Errorf("unexpected arguments: %v", remain)
	}

	sessionID := strings.TrimSpace(flagString(sessionIDFlag))
	if sessionID == "" {
		sessionID = strings.TrimSpace(os.Getenv(envSlackMsgSessionID))
	}
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "session id required")
		os.Exit(1)
	}

	dataRoot := defaultSlackLocalBotRoot()
	if _, err := lookupSessionEntry(dataRoot, sessionID); err != nil {
		if strings.Contains(err.Error(), "session not found") {
			fmt.Fprintln(os.Stderr, "session not found")
			os.Exit(1)
		}
		if strings.Contains(err.Error(), "session id required") {
			fmt.Fprintln(os.Stderr, "session id required")
			os.Exit(1)
		}
		return err
	}

	msgs, err := readSessionMessages(dataRoot, sessionID)
	if err != nil {
		return err
	}
	msgs = filterMessagesAfterID(msgs, flagString(afterMsgIDFlag))
	if limitFlag != nil && *limitFlag > 0 && len(msgs) > *limitFlag {
		// Keep the N newest messages, still printed oldest→newest.
		msgs = msgs[len(msgs)-*limitFlag:]
	}

	if jsonFlag {
		type jsonMsg struct {
			MessageID string `json:"message_id"`
			TS        string `json:"ts"`
			User      string `json:"user"`
			Text      string `json:"text"`
			Direction string `json:"direction"`
		}
		doc := struct {
			Messages []jsonMsg `json:"messages"`
		}{Messages: make([]jsonMsg, 0, len(msgs))}
		for _, m := range msgs {
			doc.Messages = append(doc.Messages, jsonMsg{
				MessageID: m.MessageID,
				TS:        m.TS,
				User:      m.User,
				Text:      m.Text,
				Direction: m.Direction,
			})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(doc); err != nil {
			return err
		}
		// Encode already appends trailing newline.
		return nil
	}

	for _, m := range msgs {
		fmt.Printf("[%s] %s: %s\n", m.TS, m.User, m.Text)
	}
	if len(msgs) == 0 {
		fmt.Println()
	}
	return nil
}

// resolveConfigPathAbs returns an absolute config path when non-empty.
func resolveConfigPathAbs(configPath string) string {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return ""
	}
	if abs, err := filepath.Abs(configPath); err == nil {
		return abs
	}
	return configPath
}
