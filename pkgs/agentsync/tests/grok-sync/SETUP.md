# Scenario

**Bug**: empty web chat when `EnsureGrokSync` gated on pre-set `runner_session_id`

```
updates.jsonl appends (turn 1, turn 2, …) or delayed grok session seed
  -> agentsync.EnsureGrokSync (single worker + optional discovery bootstrap)
  -> TailUpdatesFromOffset + Converter
  -> AppendEvent events.jsonl -> SaveCheckpoint grok-sync.json
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentsync` exports `EnsureGrokSync`,
  `StopGrokSync`, `GrokSyncWorkerCount`, `GrokSyncWorkerActive`, `GrokSyncOptions`,
  `GrokSyncState`, `GrokSyncSink`, and `NewFileGrokSyncSink`.
- Tests use isolated `t.TempDir()`; `NewFileGrokSyncSink` writes `events.jsonl`,
  `grok-sync.json`, and `meta.json` under a fake session dir.
- ACP wire line builders are inherited from this root `SETUP.md` (same shape as
  `pkgs/agenttty/tests/grok-sync-worker`).

## Steps

1. Root `Setup` applies default timing (`WorkerStartDelay`, `HoldAfterSchedule`).
2. Leaf `Setup` seeds `updates.jsonl`, optional pre-checkpoint, append schedules,
   or delayed grok session for discovery.
3. `Run` starts worker via `agentsync.EnsureGrokSync`, fires schedules, optionally
   stops/restarts for checkpoint leaves.
4. Leaf `Assert` inspects emitted events, worker count, checkpoint, `runner_session_id`.

## Context

- `grok-sync.json` path: `<sessionDir>/grok-sync.json`.
- `events.jsonl` path: `<sessionDir>/events.jsonl`.
- Write order invariant: append event → then save checkpoint (crash-safe).
- Discovery leaves start with empty `GrokSessionID` / `UpdatesPath` and rely on
  `InitialPrompt` + delayed grok session seed.

```go
import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentsync"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.WorkerStartDelay <= 0 {
		req.WorkerStartDelay = 150 * time.Millisecond
	}
	if req.HoldAfterSchedule <= 0 {
		req.HoldAfterSchedule = 800 * time.Millisecond
	}
	return nil
}

func writeUpdatesJSONL(path string, lines ...string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func appendUpdatesJSONL(path string, lines ...string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func acpUserMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpAgentMessageChunk(text string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "agent_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
	})
	return string(line)
}

func acpTurnCompleted() string {
	return `{"sessionUpdate":"turn_completed"}`
}

func encodedGrokCwd(workspace string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	return url.PathEscape(abs)
}

func grokSessionDir(grokHome, workspace, sessionUUID string) string {
	return filepath.Join(grokHome, "sessions", encodedGrokCwd(workspace), sessionUUID)
}

func grokUpdatesPath(grokHome, workspace, sessionUUID string) string {
	return filepath.Join(grokSessionDir(grokHome, workspace, sessionUUID), "updates.jsonl")
}

func grokSummaryJSON(workspace, sessionUUID string) string {
	abs, _ := filepath.Abs(workspace)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	payload := map[string]any{
		"info": map[string]any{
			"cwd":       abs,
			"sessionId": sessionUUID,
			"openedAt":  now,
		},
		"created_at": now,
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func writeFakeGrokSessionDir(t *testing.T, grokHome, workspace, sessionUUID, prompt string, initialLines ...string) string {
	t.Helper()
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir grok session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(grokSummaryJSON(workspace, sessionUUID)), 0644); err != nil {
		t.Fatalf("write summary.json: %v", err)
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	if err := appendUpdatesJSONL(updatesPath, seed...); err != nil {
		t.Fatalf("seed updates.jsonl: %v", err)
	}
	return updatesPath
}

func writeSessionMeta(sessionDir, runner, sessionID, initialPrompt, runnerSessionID string) error {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	meta := map[string]any{
		"runner":              runner,
		"session_id":          sessionID,
		"initial_prompt":      initialPrompt,
		"runner_session_id":   runnerSessionID,
		"status":              "running",
		"created_at":          now,
		"updated_at":          now,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessionDir, "meta.json"), data, 0644)
}

func readRunnerSessionID(t *testing.T, sessionDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(sessionDir, "meta.json"))
	if err != nil {
		return ""
	}
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("parse meta.json: %v", err)
	}
	id, _ := meta["runner_session_id"].(string)
	return strings.TrimSpace(id)
}

func grokSyncJSONPath(sessionDir string) string {
	return filepath.Join(sessionDir, "grok-sync.json")
}

func eventsJSONLPath(sessionDir string) string {
	return filepath.Join(sessionDir, "events.jsonl")
}

func readEventsFromSessionDir(t *testing.T, sessionDir string) []types.AgentEvent {
	t.Helper()
	data, err := os.ReadFile(eventsJSONLPath(sessionDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []types.AgentEvent
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var ev types.AgentEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event: %v\n%s", err, line)
		}
		events = append(events, ev)
	}
	return events
}

func grokTurnIndex(ev types.AgentEvent) int {
	if ev.Extensions == nil || ev.Extensions.GrokSession == nil {
		return -1
	}
	return ev.Extensions.GrokSession.TurnIndex
}

func countUserMessagesByText(events []types.AgentEvent, text string) int {
	count := 0
	for _, ev := range events {
		if ev.Type == types.ActionMessage && ev.Role == "user" && ev.Text == text {
			count++
		}
	}
	return count
}

func countActionDone(events []types.AgentEvent) int {
	count := 0
	for _, ev := range events {
		if ev.Type == types.ActionDone {
			count++
		}
	}
	return count
}

func eventsContainUserText(events []types.AgentEvent, text string) bool {
	return countUserMessagesByText(events, text) > 0
}

func readGrokSyncCheckpoint(t *testing.T, sessionDir string) agentsync.GrokSyncState {
	t.Helper()
	data, err := os.ReadFile(grokSyncJSONPath(sessionDir))
	if err != nil {
		t.Fatalf("read grok-sync.json: %v", err)
	}
	var st agentsync.GrokSyncState
	if err := json.Unmarshal(data, &st); err != nil {
		t.Fatalf("parse grok-sync.json: %v\n%s", err, string(data))
	}
	return st
}

func updatesFileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat updates.jsonl: %v", err)
	}
	return info.Size()
}

func waitForActionDoneCount(t *testing.T, sessionDir string, want int, timeout time.Duration) []types.AgentEvent {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		events := readEventsFromSessionDir(t, sessionDir)
		if countActionDone(events) >= want {
			return events
		}
		time.Sleep(50 * time.Millisecond)
	}
	return readEventsFromSessionDir(t, sessionDir)
}

func waitForDiscoveryEvents(t *testing.T, req *Request, timeout time.Duration) []types.AgentEvent {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		events := readEventsFromSessionDir(t, req.SessionDir)
		if eventsContainUserText(events, req.InitialPrompt) && countActionDone(events) >= 1 {
			return events
		}
		time.Sleep(100 * time.Millisecond)
	}
	events := readEventsFromSessionDir(t, req.SessionDir)
	t.Fatalf("timeout waiting for discovery bootstrap events; prompt=%q runner_session_id=%q events=%d",
		req.InitialPrompt, readRunnerSessionID(t, req.SessionDir), len(events))
	return events
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}
```