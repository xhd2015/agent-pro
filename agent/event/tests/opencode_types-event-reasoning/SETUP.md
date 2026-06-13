## Preconditions
- The `opencode_types` package defines `Event` and `ReasoningPart` structs.

## Steps
1. Create an `Event` with type `reasoning` and a `ReasoningPart`.
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
		Type:      "reasoning",
		SessionID: "sess_r1",
		Part: opencode_types.ReasoningPart{
			ID:   "evt_r1",
			Type: "reasoning",
			Text: "thinking step by step",
		},
	}
	data, _ := json.Marshal(evt)
	fmt.Println(string(data))
}
`
	return nil
}
```
