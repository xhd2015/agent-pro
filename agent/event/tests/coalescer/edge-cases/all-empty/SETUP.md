## Preconditions
- Sequence: `PhaseStart`(ID=x, text="") → `PhaseUpdate`(ID=x, text="") → `PhaseEnd`(ID=x, text="").

## Steps
1. Feed PhaseStart with empty text — not skipped.
2. Feed PhaseUpdate with empty text — not skipped.
3. Feed PhaseEnd with empty text — must be skipped (ID was shown via start+update).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseStart, Text: ""},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseUpdate, Text: ""},
		{Type: types.ActionMessage, ID: "x", Phase: types.PhaseEnd, Text: ""},
	}
	return nil
}
```
