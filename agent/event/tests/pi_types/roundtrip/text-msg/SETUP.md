## Preconditions
- Roundtrip: ToPi then FromPi should preserve text message content.

## Steps
1. Create an ActionMessage event with text content.
2. Call roundtrip (ToPi → FromPi) and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.Events = []types.AgentEvent{{
		Type: types.ActionMessage,
		Text: "Hello world",
	}}
	return nil
}
```
