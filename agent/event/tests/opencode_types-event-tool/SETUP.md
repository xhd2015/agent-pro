## Preconditions
- The `opencode_types` package defines `Event`, `ToolUsePart`, and `ToolUseState` structs.

## Steps
1. Create an `Event` with type `tool_use`, a `ToolUsePart` with `ToolUseState` containing bash output.
2. Marshal to JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func main() {
	evt := opencode_types.Event{
		Type:      "tool_use",
		SessionID: "sess_t1",
		Part: opencode_types.ToolUsePart{
			ID:   "evt_t1",
			Type: "tool",
			Tool: "bash",
			State: opencode_types.ToolUseState{
				Input:    map[string]any{"command": "echo hello"},
				Output:   "hello",
				ExitCode: 0,
				Status:   "completed",
			},
		},
	}
	data, _ := json.Marshal(evt)
	fmt.Println(string(data))
}
`
	return nil
}
```
