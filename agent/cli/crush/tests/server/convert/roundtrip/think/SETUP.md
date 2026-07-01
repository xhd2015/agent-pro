# Scenario

**Feature**: ToCrush/FromCrush round-trip of a single ActionThink event

## Preconditions
- Input is a single `AgentEvent` with `Type=ActionThink` and text content.

## Steps
1. Set `AgentEventsJSON` to a single think event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentEventsJSON = `[{"type":"think","text":"Let me think..."}]`
	return nil
}
```
