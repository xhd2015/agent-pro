# Scenario

**Feature**: queued user messages at `sessions/<session_id>/messages.jsonl`

```
AppendMessage(sessionID, text) -> Message{id, text, session_id, created_at}
ListMessages(sessionID) -> all queued messages in append order
PopMessages(sessionID) -> FIFO drain (oldest first), removes from queue
```

## Preconditions

- Messages persist as NDJSON lines under the flat session directory.
- APIs take bare `sessionID` only (no runner path arg).
- `PopMessages` on an empty queue returns an empty/nil slice without error.

## Steps

1. Set `req.Operation = "messages"`.
2. Leaf Setup sets `req.Action` and `req.MessageText`.
3. `Run` appends, lists, or pops as configured.
4. Leaf `Assert` checks message order, text, or empty pop.

## Context

- `Response.Messages` holds list or pop results.
- FIFO order is verified by comparing append sequence to pop sequence.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = "messages"
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	if req.SessionID == "" {
		req.SessionID = "sess_msgs"
	}
	return nil
}
```
