## Preconditions
- FromGrok: grok `text` event maps to a canonical ActionMessage.

## Steps
1. Create a GrokEvent with type `text` and data.
2. Call FromGrok and marshal the result.

```go
import (
	"testing"

	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_grok"
	req.GrokEvents = []grok_types.Event{{
		Type: grok_types.EventText,
		Data: "Here is the response",
	}}
	return nil
}
```
