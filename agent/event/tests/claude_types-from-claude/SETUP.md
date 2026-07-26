# Scenario

**Feature**: FromClaude converts a headless stream-json line sequence into canonical AgentEvents

```
# each NDJSON line unmarshals to a StreamEvent; FromClaude walks them
ClaudeInput (NDJSON) -> []claude_types.StreamEvent -> FromClaude -> []types.AgentEvent JSON

# one system init emits a step_start; assistant blocks emit message/think/tool_call; tool_result is skipped
```

## Preconditions
- `FromClaude(events []claude_types.StreamEvent, sessionID string) []types.AgentEvent` exists.
- Each leaf under this node sets `req.Target = "from_claude"` and `req.ClaudeInput` to one or more NDJSON lines.

## Steps
1. Set `req.Target = "from_claude"` so the root `Run` splits `ClaudeInput` into lines, unmarshals each into a `StreamEvent`, and calls `FromClaude`.
2. Leaf SETUPs populate `req.ClaudeInput` and `req.SessionID`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_claude"
	if req.SessionID == "" {
		req.SessionID = "sess_claude"
	}
	return nil
}
```
