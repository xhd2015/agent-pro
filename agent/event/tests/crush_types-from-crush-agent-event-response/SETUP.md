# Scenario

**Feature**: `FromCrush` parses a crush agent_event with type `response`

## Preconditions
- `FromCrush` parses a crush agent_event with type `response`.
- Response events are informational and produce no canonical ActionType (or are forwarded as-is).

## Steps
1. Construct a crush JSON event: type `agent_event`, with nested type `response`.
2. Call `FromCrush` and marshal the canonical AgentEvent as JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CrushInput = `[{
  "type": "agent_event",
  "payload": {
    "type": "response",
    "run_id": "run_001"
  }
}]`
	req.Target = "from_crush"
	req.SessionID = "sess_crush"
	return nil
}
```
