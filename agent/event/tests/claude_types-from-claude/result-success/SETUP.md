# Scenario

**Feature**: a result success line maps to one ActionDone

```
# result subtype=success terminates the run with the final text
{"type":"result","subtype":"success","is_error":false,"result":"pong"} -> ActionDone
```

## Preconditions
- `FromClaude` emits one `ActionDone` for a `result` event with `subtype:"success"` (and `is_error:false`), with `Text` = `result`.

## Steps
1. Provide one result NDJSON line with `subtype:"success"`, `is_error:false`, `result:"pong"`.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeInput = `{"type":"result","subtype":"success","is_error":false,"result":"pong","duration_ms":1,"num_turns":1,"session_id":"sess_claude","stop_reason":"end_turn"}`
	return nil
}
```
