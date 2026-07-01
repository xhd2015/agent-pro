# Scenario

**Feature**: FromCmd converts cmd session JSONL events into canonical AgentEvents

```
# each JSONL line unmarshals to a cmd Event; FromCmd walks them
CmdInput (JSONL) -> []cmd_types.Event -> FromCmd -> []types.AgentEvent JSON

# assistant reasoning → think, assistant text → message, assistant tool-call → tool_call, tool-result → output merged
```

## Preconditions
- `FromCmd(events []cmd_types.Event, sessionID string) []types.AgentEvent` exists.
- Each leaf under this node sets `req.Target = "from_cmd"` and `req.CmdInput` to one or more JSONL lines.

## Steps
1. Set `req.Target = "from_cmd"` so the root `Run` splits `CmdInput` into lines, unmarshals each into a `Event`, and calls `FromCmd`.
2. Leaf SETUPs populate `req.CmdInput` and `req.SessionID`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "from_cmd"
	if req.SessionID == "" {
		req.SessionID = "sess_cmd"
	}
	return nil
}
```
