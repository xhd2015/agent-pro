# Scenario

**Feature**: `FromCrush` parses a crush message with a tool_call part and emits ActionToolCall

## Preconditions
- `FromCrush` parses a crush message with a tool_call part and emits ActionToolCall.

## Steps
1. Construct a crush JSON event: type `message`, role `assistant`, with a `tool_call` part.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushInput = `[{
  "type": "message",
  "payload": {
    "id": "msg_3",
    "role": "assistant",
    "session_id": "sess_crush",
    "parts": [
      {"type": "tool_call", "data": {"id": "tc_1", "name": "bash", "input": "{\"command\":\"echo hi\"}"}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
