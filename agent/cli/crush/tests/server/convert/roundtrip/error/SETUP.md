# Scenario

**Feature**: ToCrush/FromCrush round-trip of a single ActionError event

## Preconditions
- Input is a single `AgentEvent` with `Type=ActionError` and error text.

## Steps
1. Set `AgentEventsJSON` to a single error event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentEventsJSON = `[{"type":"error","text":"something failed"}]`
	return nil
}
```
