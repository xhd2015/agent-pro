# Scenario

**Feature**: a user tool_result line is skipped — no canonical action is emitted

```
# tool_result echoes tool output; it folds into the preceding tool call, so FromClaude emits nothing
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"output","is_error":false}]}} -> (no action)
```

## Preconditions
- `FromClaude` emits zero `AgentEvent`s for a `user` `tool_result` block.

## Steps
1. Provide one user NDJSON line whose `message.content` has a single `tool_result` block.
2. Call `FromClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeInput = `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"output","is_error":false}]},"session_id":"sess_claude"}`
	return nil
}
```
