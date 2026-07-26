# Scenario

**Feature**: `FromCrush` parses a crush message with a text part and emits ActionMessage

## Preconditions
- `FromCrush` parses a crush message with a text part and emits ActionMessage.

## Steps
1. Construct a crush JSON event: type `message`, role `assistant`, with a `text` part.
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
    "id": "msg_2",
    "role": "assistant",
    "session_id": "sess_crush",
    "parts": [
      {"type": "text", "data": {"text": "hello world"}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
