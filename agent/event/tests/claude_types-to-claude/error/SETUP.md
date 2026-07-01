# Scenario

**Feature**: ActionError becomes one result event with subtype error

```
# canonical error -> terminal result event, subtype=error, is_error=true
ActionError{Text:"boom"} -> {"type":"result","subtype":"error","is_error":true,"result":"boom"}
```

## Preconditions
- `ToClaude` maps `ActionError` to one `result` event with `subtype:"error"`, `is_error:true`, and `result` = the event's `Text`.

## Steps
1. Provide one `ActionError` event with `Text="boom"`.
2. Call `ToClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_err_1",
		Type: types.ActionError,
		Text: "boom",
	}}
	return nil
}
```
