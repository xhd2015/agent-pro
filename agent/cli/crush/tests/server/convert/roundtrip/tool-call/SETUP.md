# Scenario

**Feature**: ToCrush/FromCrush round-trip of a single ActionToolCall event with input map

## Preconditions
- Input is a single `AgentEvent` with `Type=ActionToolCall`, a tool name, and a tool input map.

## Steps
1. Set `AgentEventsJSON` to a single tool-call event with string tool input.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AgentEventsJSON = `[{"type":"tool_call","tool":"read","tool_input":{"path":"/foo"}}]`
	return nil
}
```
