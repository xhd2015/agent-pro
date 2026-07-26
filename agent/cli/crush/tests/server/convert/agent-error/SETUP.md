# Scenario

**Feature**: UnwrapEvent on a 3-level agent_event SSE line carrying an inner error payload

## Preconditions
- Input is a valid 3-level SSE line with outer type `"agent_event"`.
- Inner payload contains an `AgentEventPayload` with `type: "error"` and an error message.

## Steps
1. Set `SSEInput` to a 3-level agent_error event.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SSEInput = `{"type":"agent_event","payload":{"type":"created","payload":{"type":"error","error":"rate limit exceeded","run_id":"run_001"}}}`
	return nil
}
```
