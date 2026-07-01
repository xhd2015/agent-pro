# Scenario

**Feature**: ActionDone becomes one result event with subtype success

```
# canonical done -> terminal result event, subtype=success, is_error=false
ActionDone{Text:"ok"} -> {"type":"result","subtype":"success","is_error":false,"result":"ok"}
```

## Preconditions
- `ToClaude` maps `ActionDone` to one `result` event with `subtype:"success"`, `is_error:false`, and `result` = the event's `Text`.

## Steps
1. Provide one `ActionDone` event with `Text="ok"`.
2. Call `ToClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_done_1",
		Type: types.ActionDone,
		Text: "ok",
	}}
	return nil
}
```
