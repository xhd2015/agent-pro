## Preconditions
- ToGrok: ActionThink maps to a grok `thought` event with the thinking text in `Data`.

## Steps
1. Create an ActionThink event with thinking text.
2. Call ToGrok and marshal the result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_grok"
	req.SessionID = "sess_test_001"
	req.Events = []types.AgentEvent{{
		Type: types.ActionThink,
		Text: "Let me think about this...",
	}}
	return nil
}
```
