# Scenario

**Feature**: ToCrush/FromCrush round-trip of a single ActionDone event

## Preconditions
- Input is a single `AgentEvent` with `Type=ActionDone` and text.

## Steps
1. Set `AgentEventsJSON` to a single done event.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AgentEventsJSON = `[{"type":"done","text":"final answer"}]`
	return nil
}
```
