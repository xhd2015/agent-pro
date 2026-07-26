# Scenario

**Feature**: ToCrush/FromCrush round-trip of a mixed-type AgentEvent array preserving count

## Preconditions
- Input is an array of `AgentEvent` with different types to verify each round-trips correctly and event count is preserved.

## Steps
1. Set `AgentEventsJSON` to a multi-event array with think, message, tool-call, and done.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AgentEventsJSON = `[{"type":"think","text":"thinking"},{"type":"message","text":"hello"},{"type":"tool_call","tool":"bash","tool_input":{"cmd":"ls"}},{"type":"done","text":"finished"}]`
	return nil
}
```
