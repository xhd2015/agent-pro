## Preconditions
- The program calls `ToCodex` with an `error` AgentEvent.

## Steps
1. Create an AgentEvent with type `error` and text.
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
		Type: types.ActionError,
		Text: "something went wrong",
	}
	result := codex.ToCodex([]types.AgentEvent{evt})
	data, _ := json.Marshal(result)
	fmt.Println(string(data))
}
`
	return nil
}
```
