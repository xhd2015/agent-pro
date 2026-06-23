# Scenario

**Bug**: `glob` AgentEvent pattern is dropped from compact trace output

```
# maintain-topic records a glob tool call while searching skill files
AgentEvent{tool=glob, tool_input.pattern=".agents/skills/git-fetch/**/*"} -> compact trace printer

# compact output keeps the search pattern visible
compact trace printer -> SEARCH block with ".agents/skills/git-fetch/**/*"
```

## Preconditions
- A `glob` tool call carries the searched pattern in `tool_input.pattern`.

## Steps
1. Build one canonical AgentEvent JSONL line for a completed `glob` tool call.
2. Include `.agents/skills/git-fetch/**/*` in `tool_input.pattern`.
3. Format the line with `print.FormatTraceLine`.

```go
import (
	"encoding/json"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	data, err := json.Marshal(types.AgentEvent{
		Type: types.ActionToolCall,
		Tool: "glob",
		ToolInput: map[string]any{
			"pattern": ".agents/skills/git-fetch/**/*",
		},
	})
	if err != nil {
		t.Fatalf("marshal glob AgentEvent: %v", err)
	}
	req.Line = string(data)
	return nil
}
```
