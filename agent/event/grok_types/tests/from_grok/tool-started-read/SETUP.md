## Preconditions
- FromGrok: grok `tool_started` with `tool_name` should map to canonical `ActionToolCall`.

## Steps
1. Create a GrokEvent with type `tool_started` and tool name in Data.
2. Call FromGrok and marshal the result.

```go
import (
	"testing"

	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_grok"
	req.GrokEvents = []grok_types.Event{{
		Type: "tool_started",
		Data: "Read",
	}}
	return nil
}
```