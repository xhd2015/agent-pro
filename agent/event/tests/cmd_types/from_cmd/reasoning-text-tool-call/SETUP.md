# Scenario

**Feature**: multiple content blocks in one assistant message produce multiple canonical events

```
# assistant message with reasoning, text, and tool-call blocks
{"role":"assistant","content":[{"type":"reasoning","text":"thinking..."},{"type":"text","text":"result"},{"type":"tool-call","toolCallId":"call_1","toolName":"bash","input":{}}]} -> ActionThink, ActionMessage, ActionToolCall
```

## Preconditions
- `FromCmd` emits events in order for each content block in the assistant message.

## Steps
1. Provide one assistant JSONL line with `reasoning`, `text`, and `tool-call` blocks.
2. Verify three canonical events in order: think, message, tool_call.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CmdInput = `{"id":"evt_4","timestamp":"2026-07-01T00:00:00Z","sessionId":"sess_cmd","parentId":"parent_4","role":"assistant","content":[{"type":"reasoning","text":"thinking..."},{"type":"text","text":"result"},{"type":"tool-call","toolCallId":"call_1","toolName":"bash","input":{}}]}`
	return nil
}
```
