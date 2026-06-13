## Preconditions
- The `opencode_types` package defines `Event` and `StepStartPart`.

## Steps
1. Parse a raw JSON string (matching real opencode step_start output) into `Event`.
2. Verify the step_start part fields are populated correctly.

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
	raw := ` + "`" + `{"type":"step_start","timestamp":1718200000123,"sessionID":"sess_ss","part":{"id":"p1","sessionID":"sess_ss","messageID":"msg_1","type":"step-start","snapshot":"snap_abc"}}` + "`" + `
	var evt opencode_types.Event
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		fmt.Println("PARSE ERROR:", err)
		return
	}
	fmt.Printf("type=%s\n", evt.Type)
	fmt.Printf("sessionID=%s\n", evt.SessionID)
	fmt.Printf("timestamp=%d\n", evt.Timestamp)

	part, ok := evt.Part.(map[string]any)
	if !ok {
		fmt.Println("PART_TYPE=map")
	}
	fmt.Printf("part.id=%v\n", part["id"])
	fmt.Printf("part.sessionID=%v\n", part["sessionID"])
	fmt.Printf("part.messageID=%v\n", part["messageID"])
	fmt.Printf("part.type=%v\n", part["type"])
	fmt.Printf("part.snapshot=%v\n", part["snapshot"])
}
`
	return nil
}
```
