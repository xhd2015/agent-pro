package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/xhd2015/agent-pro/pkgs/slackutil"
	lessflags "github.com/xhd2015/less-flags"
)

const (
	envSlackMsgSessionID = "SLACK_MSG_SESSION_ID"
	envSlackMsgConfig    = "SLACK_MSG_CONFIG"
)

const sessionHelpText = `slack-msg session: session-bound management, reply and history.

Usage:
  slack-msg session <command> [options]

Commands:
  list     List sessions from the local map
  info     Show details for one session
  update   Update session fields (e.g. workspace dir)
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
	case "list":
		return runSessionList(args[1:])
	case "info":
		return runSessionInfo(args[1:])
	case "update":
		return runSessionUpdate(args[1:])
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
	// Refresh map preview (preserve Dir via upsert or explicit pass-through).
	_ = upsertSessionEntry(dataRoot, durableSessionEntry{
		SessionID:          entry.SessionID,
		ChannelID:          entry.ChannelID,
		ThreadTS:           entry.ThreadTS,
		ConfigPath:         entry.ConfigPath,
		Dir:                entry.Dir,
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

const sessionListHelpText = `slack-msg session list: list sessions from the local map.

Usage:
  slack-msg session list [options]

Options:
  --limit N   Max rows after sort (updated_at desc)
  --json      JSON output
  -h, --help  Show help
`

const sessionInfoHelpText = `slack-msg session info: show details for one session.

Usage:
  slack-msg session info [options]

Options:
  --session-id ID   Session id (env: SLACK_MSG_SESSION_ID)
  --json            JSON output
  -h, --help        Show help
`

const sessionUpdateHelpText = `slack-msg session update: update session fields (e.g. workspace dir).

Usage:
  slack-msg session update [options]

Options:
  --session-id ID   Session id (env: SLACK_MSG_SESSION_ID)
  --dir PATH        Agent workspace directory (must exist)
  --json            JSON output
  -h, --help        Show help
`

func runSessionList(args []string) error {
	var (
		limitFlag *int
		jsonFlag  bool
	)
	remain, err := lessflags.Int("--limit", &limitFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", sessionListHelpText).
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

	dataRoot := defaultSlackLocalBotRoot()
	doc, err := loadSessionsMap(dataRoot)
	if err != nil {
		return err
	}
	entries := append([]durableSessionEntry(nil), doc.Entries...)
	sort.SliceStable(entries, func(i, j int) bool {
		return sessionUpdatedAtLess(entries[j], entries[i]) // desc
	})
	if limitFlag != nil && *limitFlag > 0 && len(entries) > *limitFlag {
		entries = entries[:*limitFlag]
	}

	if jsonFlag {
		type jsonSession struct {
			SessionID          string `json:"session_id"`
			AgentSessionID     string `json:"agent_session_id"`
			ChannelID          string `json:"channel_id"`
			ThreadTS           string `json:"thread_ts"`
			ConfigPath         string `json:"config_path"`
			Dir                string `json:"dir"`
			Kind               string `json:"kind"`
			ReplyMode          string `json:"reply_mode"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
			LastMessagePreview string `json:"last_message_preview"`
		}
		out := struct {
			Sessions []jsonSession `json:"sessions"`
		}{Sessions: make([]jsonSession, 0, len(entries))}
		for _, e := range entries {
			out.Sessions = append(out.Sessions, jsonSession{
				SessionID:          e.SessionID,
				AgentSessionID:     e.SessionID,
				ChannelID:          e.ChannelID,
				ThreadTS:           e.ThreadTS,
				ConfigPath:         e.ConfigPath,
				Dir:                e.Dir,
				Kind:               e.Kind,
				ReplyMode:          e.ReplyMode,
				CreatedAt:          e.CreatedAt,
				UpdatedAt:          e.UpdatedAt,
				LastMessagePreview: e.LastMessagePreview,
			})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		return enc.Encode(out)
	}

	if len(entries) == 0 {
		return nil
	}
	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		dir := e.Dir
		if strings.TrimSpace(dir) == "" {
			dir = "-"
		}
		rows = append(rows, []string{
			e.SessionID,
			e.ChannelID,
			dir,
			e.UpdatedAt,
			e.LastMessagePreview,
		})
	}
	fmt.Print(formatSessionListHuman(rows))
	return nil
}

// formatSessionListHuman builds the grok-style padded table.
// Columns: SESSION_ID, CHANNEL, DIR, UPDATED, PREVIEW.
func formatSessionListHuman(rows [][]string) string {
	headers := []string{"SESSION_ID", "CHANNEL", "DIR", "UPDATED", "PREVIEW"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i := 0; i < len(headers) && i < len(row); i++ {
			if n := len(row[i]); n > widths[i] {
				widths[i] = n
			}
		}
	}
	var b strings.Builder
	writeRow := func(cols []string) {
		for i := 0; i < len(headers); i++ {
			cell := ""
			if i < len(cols) {
				cell = cols[i]
			}
			if i > 0 {
				b.WriteString("  ")
			}
			if i == len(headers)-1 {
				b.WriteString(cell)
			} else {
				fmt.Fprintf(&b, "%-*s", widths[i], cell)
			}
		}
		b.WriteByte('\n')
	}
	writeRow(headers)
	for _, row := range rows {
		writeRow(row)
	}
	return b.String()
}

// sessionUpdatedAtLess reports whether a should sort before b by updated_at
// ascending (empty/missing last).
func sessionUpdatedAtLess(a, b durableSessionEntry) bool {
	ta, oka := parseSessionTime(a.UpdatedAt)
	tb, okb := parseSessionTime(b.UpdatedAt)
	if !oka && !okb {
		return a.SessionID < b.SessionID
	}
	if !oka {
		return true // empty last when sorting desc via reversed call
	}
	if !okb {
		return false
	}
	if ta.Equal(tb) {
		return a.SessionID < b.SessionID
	}
	return ta.Before(tb)
}

func parseSessionTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func resolveSessionIDFlag(flag *string) string {
	id := strings.TrimSpace(flagString(flag))
	if id == "" {
		id = strings.TrimSpace(os.Getenv(envSlackMsgSessionID))
	}
	return id
}

func runSessionInfo(args []string) error {
	var (
		sessionIDFlag *string
		jsonFlag      bool
	)
	remain, err := lessflags.String("--session-id", &sessionIDFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", sessionInfoHelpText).
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

	sessionID := resolveSessionIDFlag(sessionIDFlag)
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

	msgs, err := readSessionMessages(dataRoot, sessionID)
	if err != nil {
		return err
	}
	sessionDir := filepath.Join(dataRoot, "sessions", entry.SessionID)

	if jsonFlag {
		doc := struct {
			SessionID          string `json:"session_id"`
			AgentSessionID     string `json:"agent_session_id"`
			ChannelID          string `json:"channel_id"`
			ThreadTS           string `json:"thread_ts"`
			ConfigPath         string `json:"config_path"`
			Dir                string `json:"dir"`
			Kind               string `json:"kind"`
			ReplyMode          string `json:"reply_mode"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
			LastMessagePreview string `json:"last_message_preview"`
			MessageCount       int    `json:"message_count"`
			SessionDir         string `json:"session_dir"`
		}{
			SessionID:          entry.SessionID,
			AgentSessionID:     entry.SessionID,
			ChannelID:          entry.ChannelID,
			ThreadTS:           entry.ThreadTS,
			ConfigPath:         entry.ConfigPath,
			Dir:                entry.Dir,
			Kind:               entry.Kind,
			ReplyMode:          entry.ReplyMode,
			CreatedAt:          entry.CreatedAt,
			UpdatedAt:          entry.UpdatedAt,
			LastMessagePreview: entry.LastMessagePreview,
			MessageCount:       len(msgs),
			SessionDir:         sessionDir,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		return enc.Encode(doc)
	}

	dirHuman := entry.Dir
	if strings.TrimSpace(dirHuman) == "" {
		dirHuman = "-"
	}
	fmt.Printf("session_id: %s\n", entry.SessionID)
	fmt.Printf("agent_session_id: %s\n", entry.SessionID)
	fmt.Printf("channel_id: %s\n", entry.ChannelID)
	fmt.Printf("thread_ts: %s\n", entry.ThreadTS)
	fmt.Printf("config_path: %s\n", entry.ConfigPath)
	fmt.Printf("dir: %s\n", dirHuman)
	fmt.Printf("kind: %s\n", entry.Kind)
	fmt.Printf("reply_mode: %s\n", entry.ReplyMode)
	fmt.Printf("created_at: %s\n", entry.CreatedAt)
	fmt.Printf("updated_at: %s\n", entry.UpdatedAt)
	fmt.Printf("last_message_preview: %s\n", entry.LastMessagePreview)
	fmt.Printf("message_count: %d\n", len(msgs))
	fmt.Printf("session_dir: %s\n", sessionDir)
	return nil
}

func runSessionUpdate(args []string) error {
	var (
		sessionIDFlag *string
		dirFlag       *string
		jsonFlag      bool
	)
	remain, err := lessflags.String("--session-id", &sessionIDFlag).
		String("--dir", &dirFlag).
		Bool("--json", &jsonFlag).
		Help("-h,--help", sessionUpdateHelpText).
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

	sessionID := resolveSessionIDFlag(sessionIDFlag)
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "session id required")
		os.Exit(1)
	}

	dirArg := strings.TrimSpace(flagString(dirFlag))
	if dirArg == "" {
		fmt.Fprintln(os.Stderr, "nothing to update")
		os.Exit(1)
	}

	st, err := os.Stat(dirArg)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "dir does not exist")
			os.Exit(1)
		}
		return err
	}
	if !st.IsDir() {
		fmt.Fprintln(os.Stderr, "dir is not a directory")
		os.Exit(1)
	}
	absDir, err := filepath.Abs(dirArg)
	if err != nil {
		return err
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

	entry.Dir = absDir
	if err := upsertSessionEntry(dataRoot, *entry); err != nil {
		return err
	}
	// Re-load for accurate updated_at in output.
	updated, err := lookupSessionEntry(dataRoot, sessionID)
	if err != nil {
		return err
	}

	if jsonFlag {
		doc := struct {
			SessionID          string `json:"session_id"`
			AgentSessionID     string `json:"agent_session_id"`
			ChannelID          string `json:"channel_id"`
			ThreadTS           string `json:"thread_ts"`
			ConfigPath         string `json:"config_path"`
			Dir                string `json:"dir"`
			Kind               string `json:"kind"`
			ReplyMode          string `json:"reply_mode"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
			LastMessagePreview string `json:"last_message_preview"`
		}{
			SessionID:          updated.SessionID,
			AgentSessionID:     updated.SessionID,
			ChannelID:          updated.ChannelID,
			ThreadTS:           updated.ThreadTS,
			ConfigPath:         updated.ConfigPath,
			Dir:                updated.Dir,
			Kind:               updated.Kind,
			ReplyMode:          updated.ReplyMode,
			CreatedAt:          updated.CreatedAt,
			UpdatedAt:          updated.UpdatedAt,
			LastMessagePreview: updated.LastMessagePreview,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		return enc.Encode(doc)
	}

	fmt.Printf("OK session=%s dir=%s\n", updated.SessionID, updated.Dir)
	return nil
}
