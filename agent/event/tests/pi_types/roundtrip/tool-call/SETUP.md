# Scenario

**Feature**: Roundtrip: ToPi then FromPi should preserve tool call fields

## Preconditions
- Roundtrip: ToPi then FromPi should preserve tool call fields.

## Steps
1. Create an ActionToolCall event with tool info.
2. Call roundtrip and marshal result.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	req.Events = []types.AgentEvent{{
		Type:      types.ActionToolCall,
		Tool:      "bash",
		ToolInput: map[string]any{"command": "ls"},
		Output:    "file.txt",
	}}
	return nil
}
```
