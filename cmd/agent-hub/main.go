package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agents/agent-hub/model"
	"github.com/xhd2015/agent-pro/agents/agent-hub/storage"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-hub: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Println("Usage: agent-hub <daemon|notify|hook|fetch|replay|status|consumers|sessions|partitions|session>")
		return nil
	}
	home := hubHome()
	s := storage.New(home)
	switch args[0] {
	case "daemon":
		return runDaemon(home, args[1:])
	case "notify":
		return runNotify(s, args[1:])
	case "hook":
		return runHook(s, args[1:])
	case "fetch":
		return runFetch(s, args[1:])
	case "replay":
		return runReplay(s, args[1:])
	case "status":
		return printJSON(map[string]any{"home": home})
	case "consumers":
		return runConsumers(home)
	case "sessions":
		return runSessions(home)
	case "partitions":
		parts, err := s.Partitions()
		if err != nil {
			return err
		}
		return printJSON(map[string]any{"partitions": parts})
	case "session":
		return runSession(home, s, args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func hubHome() string {
	if home := strings.TrimSpace(os.Getenv("AGENT_HUB_HOME")); home != "" {
		return home
	}
	return ".agent-hub"
}

func runDaemon(home string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("daemon requires subcommand")
	}
	lock := filepath.Join(home, "daemon.lock")
	switch args[0] {
	case "start":
		if err := os.MkdirAll(home, 0755); err != nil {
			return err
		}
		return os.WriteFile(lock, []byte("started\n"), 0644)
	case "stop":
		if err := os.Remove(lock); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	case "status":
		_, err := os.Stat(lock)
		return printJSON(map[string]any{"running": err == nil, "home": home})
	default:
		return fmt.Errorf("unknown daemon command: %s", args[0])
	}
}

func runNotify(s *storage.Store, args []string) error {
	var data []byte
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			i++
			if i >= len(args) {
				return fmt.Errorf("--json requires value")
			}
			data = []byte(args[i])
		case "--file":
			i++
			if i >= len(args) {
				return fmt.Errorf("--file requires value")
			}
			b, err := os.ReadFile(args[i])
			if err != nil {
				return err
			}
			data = b
		}
	}
	if len(data) == 0 {
		return fmt.Errorf("notify requires --json or --file")
	}
	var event model.NormalizedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	env, err := s.Append(event, nowUTC())
	if err != nil {
		return err
	}
	return printJSON(map[string]any{"event_id": env.EventID, "partition": env.Partition, "offset": env.Offset, "received_at": env.ReceivedAt})
}

func runHook(s *storage.Store, args []string) error {
	if len(args) == 0 || args[0] != "notify" {
		return fmt.Errorf("hook requires notify")
	}
	var runner, nativeEvent string
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--runner":
			i++
			if i < len(args) {
				runner = args[i]
			}
		case "--event":
			i++
			if i < len(args) {
				nativeEvent = args[i]
			}
		}
	}
	if runner == "opencode" {
		if redirected := strings.TrimSpace(os.Getenv("AGENT_HUB_OPENCODE_RUNNER")); redirected != "" {
			runner = redirected
		}
	}
	payload, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	event, err := normalizeHook(runner, nativeEvent, payload)
	if err != nil {
		return err
	}
	env, err := s.Append(event, nowUTC())
	if err != nil {
		return err
	}
	return printJSON(map[string]any{"event_id": env.EventID, "partition": env.Partition, "offset": env.Offset})
}

func normalizeHook(runner, nativeEvent string, payload []byte) (model.NormalizedEvent, error) {
	switch runner {
	case "codex", "fake-codex", "opencode", "fake-opencode":
	default:
		return model.NormalizedEvent{}, fmt.Errorf("unknown runner: %s", runner)
	}
	eventType, ok := hookEventType(runner, nativeEvent)
	if !ok {
		return model.NormalizedEvent{}, fmt.Errorf("unknown hook event: %s", nativeEvent)
	}
	sessionID := ""
	prompt := ""
	var obj map[string]any
	if json.Unmarshal(payload, &obj) == nil {
		if v, _ := obj["session_id"].(string); v != "" {
			sessionID = v
		}
		if v, _ := obj["sessionID"].(string); v != "" {
			sessionID = v
		}
		if v, _ := obj["prompt"].(string); v != "" {
			prompt = v
		}
		if msg, _ := obj["message"].(map[string]any); msg != nil {
			if v, _ := msg["text"].(string); v != "" {
				prompt = v
			}
		}
	}
	return model.NormalizedEvent{EventType: eventType, Runner: runner, RunnerSessionID: sessionID, Prompt: prompt, Payload: payload}, nil
}

func hookEventType(runner, nativeEvent string) (model.EventType, bool) {
	switch runner {
	case "codex", "fake-codex":
		switch nativeEvent {
		case "SessionStart":
			return model.EventSessionStarted, true
		case "UserPromptSubmit":
			return model.EventPromptSubmitted, true
		case "Stop":
			return model.EventSessionFinished, true
		case "Error":
			return model.EventSessionFailed, true
		case "PreToolUse":
			return model.EventToolStarted, true
		case "PostToolUse":
			return model.EventToolFinished, true
		case "PermissionRequest":
			return model.EventPermissionAsked, true
		case "SubagentStart", "SubagentStop":
			return model.EventSessionUpdated, true
		}
	case "opencode", "fake-opencode":
		switch nativeEvent {
		case "session.created":
			return model.EventSessionStarted, true
		case "message.updated":
			return model.EventPromptSubmitted, true
		case "session.status":
			return model.EventSessionUpdated, true
		case "session.idle":
			return model.EventSessionFinished, true
		case "session.error":
			return model.EventSessionFailed, true
		case "tool.execute.before":
			return model.EventToolStarted, true
		case "tool.execute.after":
			return model.EventToolFinished, true
		case "permission.asked":
			return model.EventPermissionAsked, true
		case "permission.replied":
			return model.EventPermissionReplied, true
		}
	}
	return "", false
}

func runFetch(s *storage.Store, args []string) error {
	consumerID := ""
	limit := 1
	peek := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--consumer-id":
			i++
			if i < len(args) {
				consumerID = args[i]
			}
		case "--limit":
			i++
			if i < len(args) {
				n, err := strconv.Atoi(args[i])
				if err != nil {
					return err
				}
				limit = n
			}
		case "--peek":
			peek = true
		}
	}
	if consumerID == "" {
		return fmt.Errorf("--consumer-id is required")
	}
	if limit <= 0 {
		return fmt.Errorf("limit must be > 0")
	}
	resp, err := s.Fetch(consumerID, limit, peek)
	if err != nil {
		return err
	}
	return printJSON(resp)
}

func runReplay(s *storage.Store, args []string) error {
	consumerID := ""
	from := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--consumer-id":
			i++
			if i < len(args) {
				consumerID = args[i]
			}
		case "--from":
			i++
			if i < len(args) {
				from = args[i]
			}
		}
	}
	parts := strings.Split(from, ":")
	if consumerID == "" || len(parts) != 2 {
		return fmt.Errorf("replay requires --consumer-id and --from partition:offset in format partition:offset")
	}
	offset, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	cursor := model.Cursor{Partition: parts[0], Offset: offset}
	if err := s.SaveCursor(consumerID, cursor); err != nil {
		return err
	}
	return printJSON(map[string]any{"consumer_id": consumerID, "cursor": cursor})
}

func runConsumers(home string) error {
	dir := filepath.Join(home, "consumers")
	entries, _ := os.ReadDir(dir)
	var consumers []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".cursor.json") {
			consumers = append(consumers, strings.TrimSuffix(name, ".cursor.json"))
		}
	}
	return printJSON(map[string]any{"consumers": consumers})
}

func runSessions(home string) error {
	root := filepath.Join(home, "sessions")
	var sessions []json.RawMessage
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			sessions = append(sessions, json.RawMessage(data))
		}
		return nil
	})
	if sessions == nil {
		sessions = []json.RawMessage{}
	}
	return printJSON(map[string]any{"sessions": sessions})
}

func runSession(home string, s *storage.Store, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session requires subcommand: show, message")
	}
	switch args[0] {
	case "show":
		return runSessionShow(home, s, args[1:])
	case "message":
		return runSessionMessage(home, s, args[1:])
	default:
		return fmt.Errorf("unknown session command: %s", args[0])
	}
}

func runSessionShow(home string, s *storage.Store, args []string) error {
	runner := ""
	sessionID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runner":
			i++
			if i < len(args) {
				runner = args[i]
			}
		case "--session-id":
			i++
			if i < len(args) {
				sessionID = args[i]
			}
		}
	}
	if runner == "" {
		return fmt.Errorf("--runner is required")
	}
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}
	sd, err := s.GetSession(runner, sessionID)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("session not found: %s/%s", runner, sessionID)
	}
	return printJSON(sd)
}

func runSessionMessage(home string, s *storage.Store, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session message requires subcommand: send, list, pop")
	}
	switch args[0] {
	case "send":
		return runSessionMessageSend(home, s, args[1:])
	case "list":
		return runSessionMessageList(home, s, args[1:])
	case "pop":
		return runSessionMessagePop(home, s, args[1:])
	default:
		return fmt.Errorf("unknown session message command: %s", args[0])
	}
}

func runSessionMessageSend(home string, s *storage.Store, args []string) error {
	runner := ""
	sessionID := ""
	text := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runner":
			i++
			if i < len(args) {
				runner = args[i]
			}
		case "--session-id":
			i++
			if i < len(args) {
				sessionID = args[i]
			}
		case "--text":
			i++
			if i < len(args) {
				text = args[i]
			}
		}
	}
	if runner == "" {
		return fmt.Errorf("--runner is required")
	}
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}
	if text == "" {
		return fmt.Errorf("--text is required")
	}

	sd, err := s.GetSession(runner, sessionID)
	if err != nil {
		return err
	}
	if sd == nil {
		sd = &model.SessionData{
			Runner:          runner,
			RunnerSessionID: sessionID,
			Status:          "running",
		}
	} else if sd.Status == "completed" || sd.Status == "failed" {
		sd.Status = "running"
		sd.LastEvent = nil
	}
	if err := s.WriteSession(runner, sessionID, *sd); err != nil {
		return err
	}

	msg := model.Message{
		ID:        newMessageID(),
		Text:      text,
		SessionID: sessionID,
		CreatedAt: nowUTC(),
	}
	if err := s.AppendMessage(runner, sessionID, msg); err != nil {
		return err
	}
	return printJSON(map[string]any{
		"message":        msg,
		"session_status": sd.Status,
	})
}

func runSessionMessageList(home string, s *storage.Store, args []string) error {
	runner := ""
	sessionID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runner":
			i++
			if i < len(args) {
				runner = args[i]
			}
		case "--session-id":
			i++
			if i < len(args) {
				sessionID = args[i]
			}
		}
	}
	if runner == "" {
		return fmt.Errorf("--runner is required")
	}
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}
	sd, err := s.GetSession(runner, sessionID)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("session not found: %s/%s", runner, sessionID)
	}
	msgs, err := s.GetMessages(runner, sessionID)
	if err != nil {
		return err
	}
	if msgs == nil {
		msgs = []model.Message{}
	}
	return printJSON(map[string]any{"messages": msgs})
}

func runSessionMessagePop(home string, s *storage.Store, args []string) error {
	runner := ""
	sessionID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--runner":
			i++
			if i < len(args) {
				runner = args[i]
			}
		case "--session-id":
			i++
			if i < len(args) {
				sessionID = args[i]
			}
		}
	}
	if runner == "" {
		return fmt.Errorf("--runner is required")
	}
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}
	sd, err := s.GetSession(runner, sessionID)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("session not found: %s/%s", runner, sessionID)
	}
	msgs, err := s.GetAndClearMessages(runner, sessionID)
	if err != nil {
		return err
	}
	if msgs == nil {
		msgs = []model.Message{}
	}
	return printJSON(map[string]any{"messages": msgs})
}

func newMessageID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func printJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func nowUTC() (t time.Time) {
	return time.Now().UTC()
}
