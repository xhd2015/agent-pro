# Scenario

**Feature**: `FromCrush` handles a crush message with multiple parts (text + tool_call)

## Preconditions
- `FromCrush` handles a crush message with multiple parts (text + tool_call).
- Each part produces its own canonical AgentEvent.

## Steps
1. Construct a crush JSON event: type `message`, role `assistant`, with both `text` and `tool_call` parts.
2. Call `FromCrush` and marshal the canonical AgentEvents as JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CrushInput = `[{
  "type": "message",
  "payload": {
    "id": "msg_7",
    "role": "assistant",
    "session_id": "sess_crush",
    "parts": [
      {"type": "text", "data": {"text": "let me run that command"}},
      {"type": "tool_call", "data": {"id": "tc_2", "name": "bash", "input": "{\"command\":\"ls\"}"}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
