# Scenario

**Feature**: ActionMessage becomes one assistant event with a text block

```
# canonical message -> assistant message carrying a text content block
ActionMessage{Text:"pong"} -> {"type":"assistant","message":{"content":[{"type":"text","text":"pong"}]}}
```

## Preconditions
- `ToClaude` maps `ActionMessage` to one `assistant` event whose `message.content` has a single `text` block with the message text.

## Steps
1. Provide one `ActionMessage` event with `Text="pong"`.
2. Call `ToClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_msg_1",
		Type: types.ActionMessage,
		Text: "pong",
	}}
	return nil
}
```
