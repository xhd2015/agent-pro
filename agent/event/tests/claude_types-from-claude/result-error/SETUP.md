# Scenario

**Feature**: a result error line maps to one ActionError

```
# result subtype=error (or is_error=true) surfaces as an error action
{"type":"result","subtype":"error","is_error":true,"result":"boom"} -> ActionError
```

## Preconditions
- `FromClaude` emits one `ActionError` for a `result` event with `subtype:"error"` or `is_error:true`, with `Text` = `result`.

## Steps
1. Provide one result NDJSON line with `subtype:"error"`, `is_error:true`, `result:"boom"`.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClaudeInput = `{"type":"result","subtype":"error","is_error":true,"result":"boom","duration_ms":1,"num_turns":1,"session_id":"sess_claude"}`
	return nil
}
```
