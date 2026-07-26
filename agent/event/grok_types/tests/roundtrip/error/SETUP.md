## Preconditions
- Roundtrip: ToGrok → FromGrok preserves ActionError message.

## Steps
1. Create an ActionError event with error text.
2. Call ToGrok then FromGrok and marshal the result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "roundtrip"
	req.SessionID = "sess_rtt_001"
	req.Events = []types.AgentEvent{{
		Type: types.ActionError,
		Text: "connection refused",
	}}
	return nil
}
```
