# Scenario

**Feature**: append-only event log at `sessions/<session_id>/events.jsonl`

```
AppendEvent(sessionID, ev) -> NDJSON line appended
ReadEvents(sessionID, afterOffset) -> []AgentEvent + new byte offset
offset 0 reads from start; non-zero offset skips prior lines
```

## Preconditions

- Events are `agent/event/types.AgentEvent` serialized as one JSON object per line.
- `ReadEvents` returns `(events, nextOffset, error)` where offset is byte position for resume.
- APIs take bare `sessionID` only (no runner path arg).

## Steps

1. Set `req.Operation = "events"`.
2. Leaf Setup sets `req.Action` and `req.Events` / `req.AfterOffset` as needed.
3. `Run` appends events and reads with the configured offset.
4. Leaf `Assert` checks event count, content, and offset behavior.

## Context

- `Response.Events` is the batch returned by `ReadEvents`.
- `Response.EventsOffset` is the next read offset after the returned batch.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "events"
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	if req.SessionID == "" {
		req.SessionID = "sess_events"
	}
	return nil
}

func makeEvents(texts ...string) []types.AgentEvent {
	var out []types.AgentEvent
	for _, text := range texts {
		out = append(out, types.AgentEvent{Type: types.ActionMessage, Text: text})
	}
	return out
}
```
