## Preconditions
- The program calls `ToCodex` with an `error` AgentEvent.

## Steps
1. Create an AgentEvent with type `error` and text.
2. Call `ToCodex` and print the resulting codex events as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Events = []types.AgentEvent{{
		Type: types.ActionError,
		Text: "something went wrong",
	}}
	return nil
}
```
