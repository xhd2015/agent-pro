# Scenario

**Feature**: ToCrush/FromCrush of a tool-call event with nil ToolInput does not panic

## Preconditions
- Input is a single `AgentEvent` with `Type=ActionToolCall`, a tool name, and `ToolInput` set to `nil`.
- This verifies that nil ToolInput does not cause a panic.

## Steps
1. Set `AgentEventsJSON` to a tool-call event with no tool_input.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AgentEventsJSON = `[{"type":"tool_call","tool":"bash"}]`
	return nil
}
```
