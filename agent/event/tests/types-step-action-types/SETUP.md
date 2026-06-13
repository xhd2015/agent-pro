## Preconditions
- The `types` package defines `ActionStepStart` and `ActionStepFinish`.

## Steps
1. Create `AgentEvent` values with `step_start` and `step_finish` action types.
2. Marshal to JSON and verify the type strings.

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
	startEvt := types.AgentEvent{
		ID:        "evt_ss",
		Type:      types.ActionStepStart,
		Timestamp: 1718200000123,
	}
	startData, _ := json.Marshal(startEvt)
	fmt.Printf("step_start=%s\n", string(startData))

	finishEvt := types.AgentEvent{
		ID:        "evt_sf",
		Type:      types.ActionStepFinish,
		Timestamp: 1718200000456,
	}
	finishData, _ := json.Marshal(finishEvt)
	fmt.Printf("step_finish=%s\n", string(finishData))
}
`
	return nil
}
```
