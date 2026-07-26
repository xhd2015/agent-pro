# Scenario

**Feature**: ActionError produces message_start + message_end (with error message)

## Preconditions
- ActionError produces message_start + message_end (with error message).

## Steps
1. Create ActionError event.
2. Call ToPi and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_pi"
	req.Events = []types.AgentEvent{{
		Type: types.ActionError,
		Text: "something went wrong",
	}}
	return nil
}
```
