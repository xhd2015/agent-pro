# Scenario

**Feature**: `FromCrush` parses a crush message with a reasoning part and emits ActionThink

## Preconditions
- `FromCrush` parses a crush message with a reasoning part and emits ActionThink.

## Steps
1. Construct a crush JSON event: type `message`, role `assistant`, with a `reasoning` part.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushInput = `[{
  "type": "message",
  "payload": {
    "id": "msg_1",
    "role": "assistant",
    "session_id": "sess_crush",
    "model": "claude-sonnet",
    "provider": "anthropic",
    "parts": [
      {"type": "reasoning", "data": {"thinking": "step by step reasoning", "started_at": 1718200000}}
    ]
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
