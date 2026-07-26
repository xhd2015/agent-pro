# Scenario

**Bug**: overlapping per-message grok tails duplicate events on rapid follow-ups

```
updates.jsonl appends (turn 1, turn 2, …)
  -> EnsureGrokSync (single worker)
  -> TailUpdatesFromOffset + Converter
  -> AppendEvent events.jsonl -> SaveCheckpoint grok-sync.json
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` exports `EnsureGrokSync`,
  `StopGrokSync`, `GrokSyncWorkerCount`, `GrokSyncWorkerActive`, `GrokSyncOptions`,
  `GrokSyncState`, `GrokSyncSink`, and `NewFileGrokSyncSink`.
- Tests use isolated `t.TempDir()`; `NewFileGrokSyncSink` writes `events.jsonl` and
  `grok-sync.json` under a fake session dir.
- ACP wire line builders are inherited from this root `SETUP.md` (same shape as
  `grok-updates-tail`).

## Steps

1. Root `Setup` applies default timing (`WorkerStartDelay`, `HoldAfterSchedule`).
2. Leaf `Setup` seeds `updates.jsonl`, optional pre-checkpoint, append schedules.
3. `Run` starts worker, fires schedules, optionally stops/restarts for checkpoint leaves.
4. Leaf `Assert` inspects emitted events (read from `events.jsonl`), worker count, checkpoint.

## Context

- `grok-sync.json` path: `<sessionDir>/grok-sync.json`.
- `events.jsonl` path: `<sessionDir>/events.jsonl`.
- Write order invariant: append event → then save checkpoint (crash-safe).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agenttty"
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

func readGrokSyncCheckpoint(t *testing.T, sessionDir string) agenttty.GrokSyncState {
	t.Helper()
	data, err := os.ReadFile(grokSyncJSONPath(sessionDir))
	if err != nil {
		t.Fatalf("read grok-sync.json: %v", err)
	}
	var st agenttty.GrokSyncState
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
```