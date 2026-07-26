# Scenario

**Feature**: ToCmd converts canonical AgentEvents into cmd session events

```
# each canonical AgentEvent maps to a cmd Event
Events -> ToCmd -> []cmd_types.Event JSON

# think → reasoning block, message → text block, tool_call → tool-call block
```

## Preconditions
- `ToCmd(events []types.AgentEvent, sessionID string) []cmd_types.Event` exists.
- Each leaf under this node sets `req.Target = "cmd"` and `req.Events` to one or more `AgentEvent` values.

## Steps
1. Set `req.Target = "cmd"` so the root `Run` calls `ToCmd`.
2. Leaf SETUPs populate `req.Events` and `req.SessionID`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "cmd"
	if req.SessionID == "" {
		req.SessionID = "sess_cmd"
	}
	return nil
}
```
