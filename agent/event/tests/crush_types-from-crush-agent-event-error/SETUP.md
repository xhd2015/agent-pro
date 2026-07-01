# Scenario

**Feature**: `FromCrush` parses a crush agent_event with type `error` and emits ActionError

## Preconditions
- `FromCrush` parses a crush agent_event with type `error` and emits ActionError.

## Steps
1. Construct a crush JSON event: type `agent_event`, with nested type `error` and error text.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushInput = `[{
  "type": "agent_event",
  "payload": {
    "type": "error",
    "error": "runtime failure",
    "run_id": "run_001"
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
