# Scenario

**Feature**: ToClaude converts canonical AgentEvents back into headless stream-json events

```
# each AgentEvent becomes one StreamEvent of the matching shape
[]types.AgentEvent -> ToClaude -> []claude_types.StreamEvent -> JSON

# think->assistant(thinking), message->assistant(text), tool_call->assistant(tool_use), error->result(error), done->result(success)
```

## Preconditions
- `ToClaude(events []types.AgentEvent, sessionID string) []claude_types.StreamEvent` exists.
- Each leaf under this node sets `req.Target = "claude"` and `req.Events` to one canonical event.

## Steps
1. Set `req.Target = "claude"` so the root `Run` calls `ToClaude` and marshals the result.
2. Leaf SETUPs populate `req.Events` and `req.SessionID`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "claude"
	if req.SessionID == "" {
		req.SessionID = "sess_claude"
	}
	return nil
}
```
