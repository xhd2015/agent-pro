# Scenario

**Feature**: UnwrapEvent on a 3-level run_complete SSE line with session and text

## Preconditions
- Input is a valid 3-level SSE line with outer type `"run_complete"`.
- Inner payload contains a `RunCompletePayload` with session, text, and no error.

## Steps
1. Set `SSEInput` to a 3-level run_complete event.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SSEInput = `{"type":"run_complete","payload":{"type":"updated","payload":{"session_id":"sess_xyz789","run_id":"run_001","message_id":"msg_def456","text":"Task completed successfully","error":"","cancelled":false}}}`
	return nil
}
```
