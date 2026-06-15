## Preconditions
- `FromCrush` parses a crush run_complete event and emits ActionDone.

## Steps
1. Construct a crush JSON event: type `run_complete` with session_id, message_id, text, and no error.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushInput = `[{
  "type": "run_complete",
  "payload": {
    "session_id": "sess_crush",
    "run_id": "run_001",
    "message_id": "msg_done",
    "text": "success output"
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
