## Preconditions
- Roundtrip: ToGrok → FromGrok preserves ActionThink thinking text.

## Steps
1. Create an ActionThink event with thinking text.
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
		Type: types.ActionThink,
		Text: "Analyzing the request...",
	}}
	return nil
}
```
