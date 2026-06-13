## Preconditions
- The program calls `ToCodex` with a `think` AgentEvent.

## Steps
1. Create an AgentEvent with type `think` and text.
2. Call `ToCodex` and print the resulting codex events as JSON.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	codex "github.com/xhd2015/agent-pro/agent/event/codex_types"
)

func main() {
	evt := types.AgentEvent{
		ID:   "evt_1",
		Type: types.ActionThink,
		Text: "analyzing the request",
	}
	result := codex.ToCodex([]types.AgentEvent{evt})
	data, _ := json.Marshal(result)
	fmt.Println(string(data))
}
`
	return nil
}
```
