# Scenario

**Feature**: one assistant message with multiple content blocks emits multiple canonical actions in block order

```
# a single assistant message with text + thinking + tool_use yields message, think, tool_call
{"type":"assistant","message":{"content":[text,thinking,tool_use]}} -> ActionMessage, ActionThink, ActionToolCall
```

## Preconditions
- `FromClaude` walks `message.content` blocks in order, emitting one action per recognized block.

## Steps
1. Provide one assistant NDJSON line whose `message.content` has three blocks: `text`, `thinking`, `tool_use`.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClaudeInput = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hello"},{"type":"thinking","thinking":"reasoning"},{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"ls"}}]},"session_id":"sess_claude"}`
	return nil
}
```
