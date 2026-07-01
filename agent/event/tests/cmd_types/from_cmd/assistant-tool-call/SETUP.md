# Scenario

**Feature**: an assistant tool-call block maps to one ActionToolCall

```
# assistant message with a tool-call block
{"role":"assistant","content":[{"type":"tool-call","toolCallId":"call_1","toolName":"bash","input":{"command":"echo hi"}}]} -> ActionToolCall
```

## Preconditions
- `FromCmd` emits one `ActionToolCall` per assistant `tool-call` block, with `Tool` set to `toolName` and `ToolInput` set to `input`.

## Steps
1. Provide one assistant JSONL line with a single `tool-call` content block.
2. Call `FromCmd` via the root `Run` dispatch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CmdInput = `{"id":"evt_3","timestamp":"2026-07-01T00:00:00Z","sessionId":"sess_cmd","parentId":"parent_3","role":"assistant","content":[{"type":"tool-call","toolCallId":"call_1","toolName":"bash","input":{"command":"echo hi"}}]}`
	return nil
}
```
