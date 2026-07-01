# Scenario

**Feature**: `FromCrush` parses a crush run_complete event with an error field

## Preconditions
- `FromCrush` parses a crush run_complete event with an error field.
- Still emits ActionDone since the run itself completed.

## Steps
1. Construct a crush JSON event: type `run_complete` with a non-empty error field.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushInput = `[{
  "type": "run_complete",
  "payload": {
    "session_id": "sess_crush",
    "run_id": "run_002",
    "message_id": "msg_err",
    "error": "agent run failed"
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
