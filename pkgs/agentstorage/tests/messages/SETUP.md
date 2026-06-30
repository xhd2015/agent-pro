# Scenario

**Feature**: queued user messages at `messages.jsonl`

```
AppendMessage(text) -> Message{id, text, session_id, created_at}
ListMessages -> all queued messages in append order
PopMessages -> FIFO drain (oldest first), removes from queue
```

## Preconditions

- Messages persist as NDJSON lines under the session directory.
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

func Setup(t *testing.T, req *Request) error {
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