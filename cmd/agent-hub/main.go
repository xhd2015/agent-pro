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

	"github.com/xhd2015/agent-pro/agents/agent-hub/assets"
	"github.com/xhd2015/agent-pro/agents/agent-hub/model"
	"github.com/xhd2015/agent-pro/agents/agent-hub/storage"
)

const help = `
Usage: agent-hub <command> [ARGS]

Commands:
  daemon        manage the agent-hub daemon
  notify        send an event notification
  hook          receive a hook event from an agent runner
  fetch         fetch events for a consumer
  replay        reset a consumer cursor position
  status        show agent-hub status (home directory)
  consumers     list registered consumers
  sessions      list all sessions
  partitions    list all partitions
  session       view and manage sessions
  integration   manage agent-hub plugin integration with agent runners

Run agent-hub <command> --help for command-specific options.
`

const daemonHelp = `
Usage: agent-hub daemon <command> [ARGS]

Commands:
  start         start the daemon
  stop          stop the daemon
  status        show daemon status (running or not)

Run agent-hub daemon <command> --help for command-specific options.
`

const daemonStartHelp = `
Usage: agent-hub daemon start

Start the agent-hub daemon by creating a lock file.
`

const daemonStopHelp = `
Usage: agent-hub daemon stop

Stop the agent-hub daemon by removing the lock file.
`

const daemonStatusHelp = `
Usage: agent-hub daemon status

Show the daemon status (running or not) and home directory.
`

const notifyHelp = `
Usage: agent-hub notify --json <json> | --file <path>

Send an event notification to agent-hub.

Options:
  --json <json>   event payload as inline JSON string
  --file <path>   event payload from a JSON file
`

const hookHelp = `
Usage: agent-hub hook notify --runner <runner> --event <event>

Receive a hook event from an agent runner via stdin.

Options:
  --runner <runner>   agent runner name (e.g. opencode, codex)
  --event <event>     native event type from the runner
`

const fetchHelp = `
Usage: agent-hub fetch --consumer-id <id> [--limit <n>] [--peek]

Fetch events for a consumer.

Options:
  --consumer-id <id>   consumer identifier (required)
  --limit <n>          max events to fetch (default: 1)
  --peek               peek without advancing the cursor
`

const replayHelp = `
Usage: agent-hub replay --consumer-id <id> --from <partition:offset>

Reset a consumer cursor to replay events from a given position.

Options:
  --consumer-id <id>          consumer identifier (required)
  --from <partition:offset>   partition and offset to replay from (required)
`

const statusHelp = `
Usage: agent-hub status

Show the agent-hub home directory.
`

const consumersHelp = `
Usage: agent-hub consumers

List all registered consumers and their cursor positions.
`

const sessionsHelp = `
Usage: agent-hub sessions

List all sessions stored in agent-hub.
`

const partitionsHelp = `
Usage: agent-hub partitions

List all event partitions in agent-hub.
`

const sessionHelp = `
Usage: agent-hub session <command> [ARGS]

Commands:
  show          show session details
  message       manage session messages

Run agent-hub session <command> --help for command-specific options.
`

const sessionShowHelp = `
Usage: agent-hub session show --runner <runner> --session-id <id>

Show details for a specific session.

Options:
  --runner <runner>      agent runner name (required)
  --session-id <id>      session identifier (required)
`

const sessionMessageHelp = `
Usage: agent-hub session message <command> [ARGS]

Commands:
  send          send a message to a session
  list          list messages in a session queue
  pop           retrieve and clear messages from a session queue

Run agent-hub session message <command> --help for command-specific options.
`

const sessionMessageSendHelp = `
Usage: agent-hub session message send --runner <runner> --session-id <id> --text <text>

Send a message to a session queue.

Options:
  --runner <runner>      agent runner name (required)
  --session-id <id>      session identifier (required)
  --text <text>          message text to send (required)
`

const sessionMessageListHelp = `
Usage: agent-hub session message list --runner <runner> --session-id <id>

List messages in a session queue.

Options:
  --runner <runner>      agent runner name (required)
  --session-id <id>      session identifier (required)
`

const sessionMessagePopHelp = `
Usage: agent-hub session message pop --runner <runner> --session-id <id>

Retrieve and clear messages from a session queue.

Options:
  --runner <runner>      agent runner name (required)
  --session-id <id>      session identifier (required)
`

const integrationHelp = `
Usage: agent-hub integration <command> [ARGS]

Commands:
  status        show integration status for all runners
  opencode      install, uninstall, enable, disable the agent-hub plugin for opencode

Run agent-hub integration <command> --help for command-specific options.
`

const integrationStatusHelp = `
Usage: agent-hub integration status

Show plugin integration status for all supported runners as JSON.

Supported runners: opencode
`

const integrationOpenCodeHelp = `
Usage: agent-hub integration opencode <command> [ARGS]

Commands:
  status        show opencode plugin status
  install       install the opencode plugin
  uninstall     remove the opencode plugin
  enable        enable a disabled plugin
  disable       disable an enabled plugin

Run agent-hub integration opencode <command> --help for command-specific options.
`

const integrationOpenCodeStatusHelp = `
Usage: agent-hub integration opencode status

Show opencode plugin integration status.
`

const integrationOpenCodeInstallHelp = `
Usage: agent-hub integration opencode install [--global] [--opencode-home <dir>]

Install the agent-hub plugin for opencode. Without --global or --opencode-home,
installs to the local .opencode/plugins/ directory. With --global,
installs to ~/.config/opencode/plugins/. With --opencode-home,
installs to <dir>/plugins/. --global and --opencode-home are mutually exclusive.

Options:
  --global             install to global plugins directory
  --opencode-home <dir> install to a custom opencode home directory
`

const integrationOpenCodeUninstallHelp = `
Usage: agent-hub integration opencode uninstall [--global] [--opencode-home <dir>]

Remove the agent-hub plugin and disabled plugin file. Without
--global or --opencode-home, removes from the local .opencode/plugins/ directory.
With --global, removes from ~/.config/opencode/plugins/.
With --opencode-home, removes from <dir>/plugins/. --global and --opencode-home are mutually exclusive.

Options:
  --global             uninstall from global plugins directory
  --opencode-home <dir> uninstall from a custom opencode home directory
`

const integrationOpenCodeEnableHelp = `
Usage: agent-hub integration opencode enable [--global] [--opencode-home <dir>]

Enable a disabled plugin by renaming agent-hub.ts.disabled to
agent-hub.ts. Without --global or --opencode-home, operates on local .opencode/plugins/.
With --global, operates on ~/.config/opencode/plugins/.
With --opencode-home, operates on <dir>/plugins/. --global and --opencode-home are mutually exclusive.

Options:
  --global             enable plugin in global plugins directory
  --opencode-home <dir> enable plugin in a custom opencode home directory
`

const integrationOpenCodeDisableHelp = `
Usage: agent-hub integration opencode disable [--global] [--opencode-home <dir>]

Disable an enabled plugin by renaming agent-hub.ts to
agent-hub.ts.disabled. Without --global or --opencode-home, operates on local
.opencode/plugins/. With --global, operates on
~/.config/opencode/plugins/. With --opencode-home, operates on <dir>/plugins/.
--global and --opencode-home are mutually exclusive.

Options:
  --global             disable plugin in global plugins directory
  --opencode-home <dir> disable plugin in a custom opencode home directory
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-hub: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	switch args[0] {
	case "--help", "-h":
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	home := hubHome()
	s := storage.New(home)
	switch args[0] {
	case "daemon":
		return runDaemon(home, args[1:])
	case "notify":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(notifyHelp, "\n"))
			return nil
		}
		return runNotify(s, args[1:])
	case "hook":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(hookHelp, "\n"))
			return nil
		}
		return runHook(s, args[1:])
	case "fetch":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(fetchHelp, "\n"))
			return nil
		}
		return runFetch(s, args[1:])
	case "replay":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(replayHelp, "\n"))
			return nil
		}
		return runReplay(s, args[1:])
	case "status":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(statusHelp, "\n"))
			return nil
		}
		return printJSON(map[string]any{"home": home})
	case "consumers":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(consumersHelp, "\n"))
			return nil
		}
		return runConsumers(home)
	case "sessions":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(sessionsHelp, "\n"))
			return nil
		}
		return runSessions(home)
	case "partitions":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(partitionsHelp, "\n"))
			return nil
		}
		parts, err := s.Partitions()
		if err != nil {
			return err
		}
		return printJSON(map[string]any{"partitions": parts})
	case "session":
		return runSession(home, s, args[1:])
	case "integration":
		return runIntegration(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func hubHome() string {
	if home := strings.TrimSpace(os.Getenv("AGENT_HUB_HOME")); home != "" {
		return home
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return ".agent-hub"
	}
	return filepath.Join(userHome, ".agent-hub")
}

func runDaemon(home string, args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(daemonHelp, "\n"))
		return nil
	}
	lock := filepath.Join(home, "daemon.lock")
	switch args[0] {
	case "start":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(daemonStartHelp, "\n"))
			return nil
		}
		if err := os.MkdirAll(home, 0755); err != nil {
			return err
		}
		return os.WriteFile(lock, []byte("started\n"), 0644)
	case "stop":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(daemonStopHelp, "\n"))
			return nil
		}
		if err := os.Remove(lock); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	case "status":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(daemonStatusHelp, "\n"))
			return nil
		}
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
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(sessionHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "show":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(sessionShowHelp, "\n"))
			return nil
		}
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
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(sessionMessageHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "send":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(sessionMessageSendHelp, "\n"))
			return nil
		}
		return runSessionMessageSend(home, s, args[1:])
	case "list":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(sessionMessageListHelp, "\n"))
			return nil
		}
		return runSessionMessageList(home, s, args[1:])
	case "pop":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(sessionMessagePopHelp, "\n"))
			return nil
		}
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

func runIntegration(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(integrationHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "status":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(integrationStatusHelp, "\n"))
			return nil
		}
		return runIntegrationStatus()
	case "opencode":
		return runIntegrationOpenCode(args[1:])
	default:
		return fmt.Errorf("unknown runner: %s (supported: opencode)", args[0])
	}
}

func runIntegrationOpenCode(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(integrationOpenCodeHelp, "\n"))
		return nil
	}
	switch args[0] {
	case "status":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(integrationOpenCodeStatusHelp, "\n"))
			return nil
		}
		status, err := integrationStatusForOpenCode()
		if err != nil {
			return err
		}
		return printJSON(status)
	case "install":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(integrationOpenCodeInstallHelp, "\n"))
			return nil
		}
		return runIntegrationOpenCodeInstall(args[1:])
	case "uninstall":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(integrationOpenCodeUninstallHelp, "\n"))
			return nil
		}
		return runIntegrationOpenCodeUninstall(args[1:])
	case "enable":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(integrationOpenCodeEnableHelp, "\n"))
			return nil
		}
		return runIntegrationOpenCodeEnable(args[1:])
	case "disable":
		if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
			fmt.Print(strings.TrimPrefix(integrationOpenCodeDisableHelp, "\n"))
			return nil
		}
		return runIntegrationOpenCodeDisable(args[1:])
	default:
		return fmt.Errorf("unknown opencode command: %s", args[0])
	}
}

func integrationPluginsDir(global bool, opencodeHome string) (string, error) {
	if global && opencodeHome != "" {
		return "", fmt.Errorf("--global and --opencode-home are mutually exclusive")
	}
	if opencodeHome != "" {
		return filepath.Join(opencodeHome, "plugins"), nil
	}
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find home directory: %w", err)
		}
		return filepath.Join(home, ".config", "opencode", "plugins"), nil
	}
	return integrationLocalPluginsDir()
}

func parseIntegrationFlags(args []string) (global bool, opencodeHome string, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--global":
			global = true
		case "--opencode-home":
			i++
			if i >= len(args) {
				return false, "", fmt.Errorf("--opencode-home requires a value")
			}
			opencodeHome = args[i]
		}
	}
	return global, opencodeHome, nil
}

func integrationLocalPluginsDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}
	return filepath.Join(cwd, ".opencode", "plugins"), nil
}

type integrationStatus struct {
	Installed bool   `json:"installed"`
	Enabled   bool   `json:"enabled"`
	Path      string `json:"path"`
}

func integrationStatusForOpenCode() (integrationStatus, error) {
	localDir, err := integrationLocalPluginsDir()
	if err != nil {
		return integrationStatus{}, err
	}
	tsPath := filepath.Join(localDir, "agent-hub.ts")
	disabledPath := filepath.Join(localDir, "agent-hub.ts.disabled")
	_, tserr := os.Stat(tsPath)
	_, diserr := os.Stat(disabledPath)
	if tserr == nil {
		return integrationStatus{Installed: true, Enabled: true, Path: tsPath}, nil
	}
	if diserr == nil {
		return integrationStatus{Installed: true, Enabled: false, Path: disabledPath}, nil
	}
	return integrationStatus{Installed: false, Enabled: false, Path: tsPath}, nil
}

func runIntegrationStatus() error {
	status, err := integrationStatusForOpenCode()
	if err != nil {
		return err
	}
	return printJSON(map[string]any{"opencode": status})
}

func runIntegrationOpenCodeInstall(args []string) error {
	globalFlag, opencodeHome, err := parseIntegrationFlags(args)
	if err != nil {
		return err
	}
	pluginsDir, err := integrationPluginsDir(globalFlag, opencodeHome)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("create plugins directory: %w", err)
	}
	dstPath := filepath.Join(pluginsDir, "agent-hub.ts")
	if err := os.WriteFile(dstPath, []byte(assets.AgentHubPlugin), 0644); err != nil {
		return fmt.Errorf("write plugin file: %w", err)
	}
	fmt.Printf("Installed: %s\n", dstPath)
	return nil
}

func runIntegrationOpenCodeUninstall(args []string) error {
	globalFlag, opencodeHome, err := parseIntegrationFlags(args)
	if err != nil {
		return err
	}
	pluginsDir, err := integrationPluginsDir(globalFlag, opencodeHome)
	if err != nil {
		return err
	}
	tsPath := filepath.Join(pluginsDir, "agent-hub.ts")
	disabledPath := filepath.Join(pluginsDir, "agent-hub.ts.disabled")

	removed := false
	if err := os.Remove(tsPath); err == nil {
		removed = true
	}
	if err := os.Remove(disabledPath); err == nil {
		removed = true
	}
	if !removed {
		return fmt.Errorf("plugin not installed")
	}
	fmt.Println("Uninstalled")
	return nil
}

func runIntegrationOpenCodeEnable(args []string) error {
	globalFlag, opencodeHome, err := parseIntegrationFlags(args)
	if err != nil {
		return err
	}
	pluginsDir, err := integrationPluginsDir(globalFlag, opencodeHome)
	if err != nil {
		return err
	}
	tsPath := filepath.Join(pluginsDir, "agent-hub.ts")
	disabledPath := filepath.Join(pluginsDir, "agent-hub.ts.disabled")

	if _, err := os.Stat(tsPath); err == nil {
		fmt.Println("already enabled")
		return nil
	}
	if _, err := os.Stat(disabledPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin not installed")
	}
	if err := os.Rename(disabledPath, tsPath); err != nil {
		return fmt.Errorf("rename plugin: %w", err)
	}
	fmt.Println("Enabled")
	return nil
}

func runIntegrationOpenCodeDisable(args []string) error {
	globalFlag, opencodeHome, err := parseIntegrationFlags(args)
	if err != nil {
		return err
	}
	pluginsDir, err := integrationPluginsDir(globalFlag, opencodeHome)
	if err != nil {
		return err
	}
	tsPath := filepath.Join(pluginsDir, "agent-hub.ts")
	disabledPath := filepath.Join(pluginsDir, "agent-hub.ts.disabled")

	if _, err := os.Stat(disabledPath); err == nil {
		fmt.Println("already disabled")
		return nil
	}
	if _, err := os.Stat(tsPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin not installed")
	}
	if err := os.Rename(tsPath, disabledPath); err != nil {
		return fmt.Errorf("rename plugin: %w", err)
	}
	fmt.Println("Disabled")
	return nil
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
