## Preconditions
- Roundtrip: ToGrok → FromGrok preserves ActionMessage text content.

## Steps
1. Create an ActionMessage event with text.
2. Call ToGrok then FromGrok and marshal the result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.SessionID = "sess_rtt_001"
	req.Events = []types.AgentEvent{{
		Type: types.ActionMessage,
		Text: "Hello world",
	}}
	return nil
}
```
