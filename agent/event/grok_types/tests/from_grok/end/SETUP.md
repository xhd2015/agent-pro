## Preconditions
- FromGrok: grok `end` event maps to ActionDone, with sessionId captured in ToolInput.

## Steps
1. Create a GrokEvent with type `end` and session ID.
2. Call FromGrok and marshal the result.

```go
import (
	"testing"

	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_grok"
	req.GrokEvents = []grok_types.Event{{
		Type:      grok_types.EventEnd,
		SessionID: "grok_session_42",
	}}
	return nil
}
```
