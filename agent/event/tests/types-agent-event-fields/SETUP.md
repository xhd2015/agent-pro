## Preconditions
- The program imports `agent/event/types` and marshals an `AgentEvent` with all fields populated.

## Steps
1. Create an AgentEvent with every field set.
2. Marshal to JSON and print.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func main() {
	ec := 42
	evt := types.AgentEvent{
		ID:        "evt_001",
		Type:      types.ActionToolCall,
		Text:      "hello world",
		Tool:      "bash",
		ToolInput: map[string]any{"command": "echo hi"},
		Output:    "hi",
		Stderr:    "err msg",
		ExitCode:  &ec,
		Changes:   []types.FileChange{{Path: "foo.txt", Kind: "add"}},
	}
	data, _ := json.Marshal(evt)
	fmt.Println(string(data))
}
`
	return nil
}
```
