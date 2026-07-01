# Scenario

**Feature**: a tool-call followed by a tool-result merges output into the tool_call event

```
# assistant tool-call, then tool result with output
{"role":"assistant","content":[{"type":"tool-call","toolCallId":"call_1","toolName":"bash","input":{"command":"echo hi"}}]}
{"role":"tool","content":[{"type":"tool-result","toolCallId":"call_1","toolName":"bash","output":{"type":"text","value":"hi"}}]}
-> ActionToolCall with Output="hi"
```

## Preconditions
- `FromCmd` tracks the latest tool-call and merges the matching tool-result's output into the canonical event's `output` field.

## Steps
1. Provide two JSONL lines: assistant tool-call, then tool result.
2. Verify one ActionToolCall with output set to `hi`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CmdInput = `{"id":"evt_5","timestamp":"2026-07-01T00:00:00Z","sessionId":"sess_cmd","parentId":"parent_5","role":"assistant","content":[{"type":"tool-call","toolCallId":"call_1","toolName":"bash","input":{"command":"echo hi"}}]}
{"id":"evt_6","timestamp":"2026-07-01T00:00:01Z","sessionId":"sess_cmd","parentId":"parent_6","role":"tool","content":[{"type":"tool-result","toolCallId":"call_1","toolName":"bash","output":{"type":"text","value":"hi"}}]}`
	return nil
}
```
