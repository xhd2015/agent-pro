# Scenario

**Feature**: background reconcile heals empty events with per-session flock

```
ReconcileOnce scans meta.json candidates
  -> skip if worker active
  -> TryAcquireSessionLock (LOCK_NB)
  -> EnsureGrokSync with discovery bootstrap
  -> events.jsonl populated
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentsync` exports `TryAcquireSessionLock`,
  `ReconcileOnce`, `ReconcileOptions`, and re-exports worker APIs from grok sync.
- Tests use isolated `t.TempDir()` with `AGENT_RUN_HOME` session layout.
- ACP builders shared with `grok-sync` tree shape (reimplemented here — separate DOCTEST root).

## Steps

1. Root `Setup` sets default runner/session ids.
2. Leaf `Setup` seeds session meta, grok updates, or lock dir as needed.
3. `Run` executes flock or reconcile mode.
4. Leaf `Assert` checks lock semantics or event counts.

## Context

- Lock file: `<sessionDir>/grok-sync.lock`.
- Reconcile target: `sessions/grok-tty/<id>/meta.json` with `initial_prompt`.

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
)

const (
	reconcileHealPrompt   = "reconcile heal probe prompt"
	reconcileHealGrokUUID = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	reconcileHealReply    = "reconcile-heal-reply-marker"

	reconcileSkipPrompt   = "reconcile skip worker prompt"
	reconcileSkipGrokUUID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	reconcileSkipReply    = "reconcile-skip-reply-marker"
)

func Setup(t *testing.T, req *Request) error {
	if req.ReconcileTimeout <= 0 {
		req.ReconcileTimeout = 10 * time.Second
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

func appendUpdatesJSONL(path string, lines ...string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
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

func writeFakeGrokSessionDir(grokHome, workspace, sessionUUID, prompt string, initialLines ...string) (string, error) {
	dir := grokSessionDir(grokHome, workspace, sessionUUID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(grokSummaryJSON(workspace, sessionUUID)), 0644); err != nil {
		return "", err
	}
	updatesPath := filepath.Join(dir, "updates.jsonl")
	seed := []string{acpUserMessageChunk(prompt)}
	seed = append(seed, initialLines...)
	if err := appendUpdatesJSONL(updatesPath, seed...); err != nil {
		return "", err
	}
	return updatesPath, nil
}

func writeSessionMeta(sessionDir, runner, sessionID, initialPrompt, runnerSessionID, status string) error {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	meta := map[string]any{
		"runner":            runner,
		"session_id":        sessionID,
		"initial_prompt":    initialPrompt,
		"runner_session_id": runnerSessionID,
		"status":            status,
		"workspace":         sessionDir,
		"created_at":        now,
		"updated_at":        now,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessionDir, "meta.json"), data, 0644)
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

func countUserMessagesByText(events []types.AgentEvent, text string) int {
	count := 0
	for _, ev := range events {
		if ev.Type == types.ActionMessage && ev.Role == "user" && ev.Text == text {
			count++
		}
	}
	return count
}

func seedFinishedEmptySession(t *testing.T, req *Request, status string) error {
	t.Helper()
	prompt := req.InitialPrompt
	if prompt == "" {
		prompt = reconcileHealPrompt
	}
	grokID := req.GrokSessionID
	if grokID == "" {
		grokID = reconcileHealGrokUUID
	}
	updatesPath, err := writeFakeGrokSessionDir(req.GrokHome, req.Workspace, grokID, prompt,
		acpAgentMessageChunk(reconcileHealReply),
		acpTurnCompleted(),
	)
	if err != nil {
		return err
	}
	req.GrokSessionID = grokID
	req.GrokUpdatesPath = updatesPath
	return writeSessionMeta(req.SessionDir, req.Runner, req.SessionID, prompt, "", status)
}

func seedRunningSessionWithUpdates(t *testing.T, req *Request) error {
	t.Helper()
	prompt := req.InitialPrompt
	if prompt == "" {
		prompt = reconcileSkipPrompt
	}
	grokID := req.GrokSessionID
	if grokID == "" {
		grokID = reconcileSkipGrokUUID
	}
	updatesPath, err := writeFakeGrokSessionDir(req.GrokHome, req.Workspace, grokID, prompt,
		acpAgentMessageChunk(reconcileSkipReply),
		acpTurnCompleted(),
	)
	if err != nil {
		return err
	}
	req.GrokSessionID = grokID
	req.GrokUpdatesPath = updatesPath
	if err := writeSessionMeta(req.SessionDir, req.Runner, req.SessionID, prompt, grokID, "running"); err != nil {
		return err
	}
	return nil
}
```