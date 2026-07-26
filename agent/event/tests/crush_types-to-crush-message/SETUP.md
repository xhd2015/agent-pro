# Scenario

**Feature**: `ToCrush` converts ActionMessage to a crush message with a text part

## Preconditions
- `ToCrush` converts ActionMessage to a crush message with a text part.

## Steps
1. Create an AgentEvent with type `message` and text content.
2. Call `ToCrush` and marshal the result as JSON.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Events = []types.AgentEvent{{
		ID:   "evt_msg_1",
		Type: types.ActionMessage,
		Text: "here is the result",
	}}
	req.Target = "crush"
	req.SessionID = "sess_crush"
	return nil
}
```
