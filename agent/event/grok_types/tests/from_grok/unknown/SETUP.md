## Preconditions
- FromGrok: grok events with unknown `type` values should be skipped.

## Steps
1. Create a GrokEvent with an unrecognized type like `progress`.
2. Call FromGrok and marshal the result.
3. Expect empty JSON array (no agent events).

```go
import (
	"testing"

	grok_types "github.com/xhd2015/agent-pro/agent/event/grok_types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_grok"
	req.GrokEvents = []grok_types.Event{{
		Type: "progress",
		Data: "50% complete",
	}}
	return nil
}
```
