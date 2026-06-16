## Preconditions
- Roundtrip: ToGrok → FromGrok preserves ActionDone and session ID.

## Steps
1. Create an ActionDone event.
2. Call ToGrok with session ID, then FromGrok, and marshal the result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.SessionID = "sess_xyz_789"
	req.Events = []types.AgentEvent{{
		Type: types.ActionDone,
	}}
	return nil
}
```
