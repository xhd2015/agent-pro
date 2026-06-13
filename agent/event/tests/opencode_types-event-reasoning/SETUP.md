## Preconditions
- The `opencode_types` package defines `Event` and `ReasoningPart` structs.

## Steps
1. Create an `Event` with type `reasoning` and a `ReasoningPart`.
2. Marshal to JSON.

```go
import (
	"testing"

	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Value = opencode_types.Event{
		Type:      "reasoning",
		SessionID: "sess_r1",
		Part: opencode_types.ReasoningPart{
			ID:   "evt_r1",
			Type: "reasoning",
			Text: "thinking step by step",
		},
	}
	return nil
}
```
