## Preconditions
- `FromCrush` parses a crush run_complete event with cancelled=true.
- Emits ActionDone with cancelled information.

## Steps
1. Construct a crush JSON event: type `run_complete` with cancelled=true.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushInput = `[{
  "type": "run_complete",
  "payload": {
    "session_id": "sess_crush",
    "run_id": "run_003",
    "message_id": "msg_cancel",
    "cancelled": true
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
