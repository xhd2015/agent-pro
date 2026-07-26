# Scenario

**Feature**: `FromCrush` parses a crush message with a tool_result part

## Preconditions
- `FromCrush` parses a crush message with a tool_result part.
- Tool results are informational only and produce no canonical ActionType.

## Steps
1. Construct a crush JSON event: type `message`, role `assistant`, with a `tool_result` part.
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
    "id": "msg_4",
    "role": "assistant",
    "session_id": "sess_crush",
    "parts": [
      {"type": "tool_result", "data": {"tool_call_id": "tc_1", "name": "bash", "content": "hi"}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
