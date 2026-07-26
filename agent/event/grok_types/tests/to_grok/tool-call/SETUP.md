## Preconditions
- ToGrok: ActionToolCall is not a grok native event type and should be skipped.

## Steps
1. Create an ActionToolCall event with tool info.
2. Call ToGrok and marshal the result.
3. Expect empty JSON array (no events).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_grok"
	req.SessionID = "sess_test_001"
	req.Events = []types.AgentEvent{{
		Type: types.ActionToolCall,
		Tool: "bash",
		ToolInput: map[string]any{
			"command": "ls",
		},
	}}
	return nil
}
```
