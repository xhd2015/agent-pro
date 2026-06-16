## Preconditions
- FromGrok: grok `thought` event maps to a canonical ActionThink.

## Steps
1. Create a GrokEvent with type `thought` and data.
2. Call FromGrok and marshal the result.

```go
import (
	"testing"

	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_grok"
	req.GrokEvents = []grok_types.Event{{
		Type: grok_types.EventThought,
		Data: "I need to consider the user's request",
	}}
	return nil
}
```
