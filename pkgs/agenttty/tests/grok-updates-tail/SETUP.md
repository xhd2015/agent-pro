# Scenario

**Feature**: direct unit harness for `TailUpdatesFromOffset` on synthetic `updates.jsonl`

```
write updates.jsonl (seed + scheduled appends)
  -> TailUpdatesFromOffset(ctx, path, startOffset, emit)
  -> collect []types.AgentEvent until ctx cancel
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` exports `TailUpdatesFromOffset`.
- Tests use isolated `t.TempDir()`; no real grok binary or agent-run CLI.
- ACP wire line builders live in this root `SETUP.md` and are inherited by all leaves.

## Steps

1. Root `Setup` is a no-op pass-through (per-leaf `Setup` configures `Request`).
2. `Run` seeds `updates.jsonl`, starts tail goroutine, fires `AppendSchedules`, cancels ctx.
3. Leaf `Assert` inspects collected `Response.Events` / `EventTexts`.

## Context

- `TailStartDelay` (default 150ms) allows bootstrap read before scheduled appends.
- `HoldAfterSchedule` (default 600ms) keeps tail alive after last append so WatchLine can deliver.
- `StartOffsetAtEOF` sets offset to current file size (stale-skip semantics).

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
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TailStartDelay <= 0 {
		req.TailStartDelay = 150 * time.Millisecond
	}
	if req.HoldAfterSchedule <= 0 {
		req.HoldAfterSchedule = 600 * time.Millisecond
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

func acpToolCall(toolCallID, kind, title string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call",
		"toolCallId":    toolCallID,
		"kind":          kind,
		"title":         title,
		"status":        "pending",
	})
	return string(line)
}

func acpToolCallUpdate(toolCallID, status, output string) string {
	line, _ := json.Marshal(map[string]any{
		"sessionUpdate": "tool_call_update",
		"toolCallId":    toolCallID,
		"status":        status,
		"content": []map[string]any{
			{
				"type": "content",
				"content": map[string]any{
					"type": "text",
					"text": output,
				},
			},
		},
	})
	return string(line)
}

func acpTurnCompleted() string {
	return `{"sessionUpdate":"turn_completed"}`
}

func eventsContainText(events []string, want string) bool {
	want = strings.ToLower(want)
	for _, text := range events {
		if strings.Contains(strings.ToLower(text), want) {
			return true
		}
	}
	return false
}

func eventsContainActionDone(t *testing.T, events []types.AgentEvent) bool {
	t.Helper()
	for _, ev := range events {
		if ev.Type == types.ActionDone {
			return true
		}
	}
	return false
}
```