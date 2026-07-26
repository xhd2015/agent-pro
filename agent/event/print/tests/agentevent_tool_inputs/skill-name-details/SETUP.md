# Scenario

**Bug**: canonical `skill` AgentEvent tool calls print only the `SKILL` header

```
# maintain-topic records skill invocations as canonical AgentEvent tool_call lines
AgentEvent{tool=skill, tool_input.name=confluence-fetch} -> compact trace printer

# compact output should include the selected skill name below the SKILL header
compact trace printer -> SKILL block with confluence-fetch
```

## Preconditions
- A `skill` tool call carries the selected skill name in `tool_input.name`.
- This matches maintain-topic sessions that install and invoke `confluence-fetch`
  and `git-fetch` via the skill hub.

## Steps
1. Build one canonical AgentEvent JSONL line for a completed `skill` tool call.
2. Include `confluence-fetch` in `tool_input.name`.
3. Format the line with `print.FormatTraceLine`.

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	data, err := json.Marshal(types.AgentEvent{
		Type: types.ActionToolCall,
		Tool: "skill",
		ToolInput: map[string]any{
			"name": "confluence-fetch",
		},
	})
	if err != nil {
		t.Fatalf("marshal skill AgentEvent: %v", err)
	}
	req.Line = string(data)
	return nil
}
```