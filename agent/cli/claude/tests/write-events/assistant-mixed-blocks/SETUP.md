# Scenario

**Feature**: One assistant line with text+thinking+tool_use blocks emits three AgentEvents in array order

```
# assistant content blocks iterate in order: text -> message, thinking -> think, tool_use -> tool_call
assistant-mixed-blocks -> WriteClaudeLine(assistant text+thinking+tool_use)
WriteClaudeLine -> RawLog (3 AgentEvents: message, think, tool_call)
```

## Preconditions
- `claude_types.FromClaude` iterates `assistant.message.content` in array order,
  emitting `ActionMessage` (text), `ActionThink` (thinking), `ActionToolCall`
  (tool_use).

## Steps
1. Set `req.ClaudeLines` to one `assistant` line with three content blocks.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeLines = []string{
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hi"},{"type":"thinking","thinking":"why"},{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"ls"}}]},"session_id":"sess"}`,
	}
	return nil
}
```
