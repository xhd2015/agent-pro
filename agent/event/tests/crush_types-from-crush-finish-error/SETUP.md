# Scenario

**Feature**: `FromCrush` parses a crush message with a finish part having reason `error`

## Preconditions
- `FromCrush` parses a crush message with a finish part having reason `error`.
- This should emit ActionError.

## Steps
1. Construct a crush JSON event: type `message`, role `assistant`, with a `finish` part reason=error.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CrushInput = `[{
  "type": "message",
  "payload": {
    "id": "msg_5",
    "role": "assistant",
    "session_id": "sess_crush",
    "parts": [
      {"type": "finish", "data": {"reason": "error", "time": 1718200000, "message": "generation failed"}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
