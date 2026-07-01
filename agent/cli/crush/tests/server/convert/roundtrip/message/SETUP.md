# Scenario

**Feature**: ToCrush/FromCrush round-trip of a single ActionMessage event

## Preconditions
- Input is a single `AgentEvent` with `Type=ActionMessage` and text content.

## Steps
1. Set `AgentEventsJSON` to a single message event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentEventsJSON = `[{"type":"message","text":"Hello world"}]`
	return nil
}
```
