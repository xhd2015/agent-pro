## Preconditions
- ToGrok: ActionError maps to a grok `text` event with the error message in `Data`.

## Steps
1. Create an ActionError event with error text.
2. Call ToGrok and marshal the result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_grok"
	req.SessionID = "sess_test_001"
	req.Events = []types.AgentEvent{{
		Type: types.ActionError,
		Text: "something went wrong",
	}}
	return nil
}
```
