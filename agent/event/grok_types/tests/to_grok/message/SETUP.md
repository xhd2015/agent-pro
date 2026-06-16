## Preconditions
- ToGrok: ActionMessage maps to a grok `text` event with the message text in `Data`.

## Steps
1. Create an ActionMessage event with text content.
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
		Type: types.ActionMessage,
		Text: "Hello world",
	}}
	return nil
}
```
