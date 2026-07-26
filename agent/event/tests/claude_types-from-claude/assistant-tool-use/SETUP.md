# Scenario

**Feature**: an assistant tool_use block maps to one ActionToolCall

```
# tool_use block name+input become Tool and ToolInput
{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"ls"}}]}} -> ActionToolCall
```

## Preconditions
- `FromClaude` emits one `ActionToolCall` per assistant `tool_use` block, with `Tool` = block `name` and `ToolInput` = block `input`.

## Steps
1. Provide one assistant NDJSON line whose `message.content` has a single `tool_use` block (`name=Bash`, `input={"command":"ls"}`).
2. Call `FromClaude` via the root `Run` dispatch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClaudeInput = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"ls"}}]},"session_id":"sess_claude"}`
	return nil
}
```
