## Preconditions
- ToGrok: ActionDone maps to a grok `end` event with the provided session ID.

## Steps
1. Create an ActionDone event.
2. Call ToGrok with a session ID and marshal the result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_grok"
	req.SessionID = "sess_abc_123"
	req.Events = []types.AgentEvent{{
		Type: types.ActionDone,
	}}
	return nil
}
```
