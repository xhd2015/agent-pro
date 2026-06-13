## Preconditions
- The program calls `ToOpencode` with a `think` AgentEvent.

## Steps
1. Create an AgentEvent with type `think` and text.
2. Call `ToOpencode` and print the resulting opencode events as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	opencode "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func main() {
	evt := types.AgentEvent{
		ID:   "evt_1",
		Type: types.ActionThink,
		Text: "thinking about the problem",
	}
	result := opencode.ToOpencode([]types.AgentEvent{evt}, "sess_001")
	data, _ := json.Marshal(result)
	fmt.Println(string(data))
}
`
	return nil
}
```
