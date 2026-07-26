# Scenario

**Feature**: `FromCrush` skips messages with role `user` or `system` — only `assistant` messages are converted

## Preconditions
- `FromCrush` skips messages with role `user` or `system` — only `assistant` messages are converted.

## Steps
1. Construct a crush JSON event: type `message`, role `user`, with a text part.
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
    "id": "msg_8",
    "role": "user",
    "session_id": "sess_crush",
    "parts": [
      {"type": "text", "data": {"text": "what is the answer?"}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
