# Scenario

**Feature**: ActionThink becomes one assistant event with a thinking block

```
# canonical think -> assistant message carrying a thinking content block
ActionThink{Text:"reasoning"} -> {"type":"assistant","message":{"content":[{"type":"thinking","thinking":"reasoning"}]}}
```

## Preconditions
- `ToClaude` maps `ActionThink` to one `assistant` event whose `message.content` has a single `thinking` block with the thinking text.

## Steps
1. Provide one `ActionThink` event with `Text="reasoning"`.
2. Call `ToClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_think_1",
		Type: types.ActionThink,
		Text: "reasoning",
	}}
	return nil
}
```
