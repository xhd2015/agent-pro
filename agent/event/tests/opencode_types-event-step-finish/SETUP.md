## Preconditions
- The `opencode_types` package defines `Event`, `StepFinishPart`, `Tokens`, and `CacheTokens`.

## Steps
1. Parse a raw JSON string (matching real opencode step_finish output) into `Event`.
2. Verify nested fields like reason, cost, and token breakdown are populated.

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
	raw := ` + "`" + `{"type":"step_finish","timestamp":1718200000456,"sessionID":"sess_sf","part":{"id":"p2","sessionID":"sess_sf","messageID":"msg_2","type":"step-finish","reason":"stop","snapshot":"snap_xyz","cost":0.015,"tokens":{"input":120,"output":80,"reasoning":40,"cache":{"read":10,"write":5}}}}` + "`" + `
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
	fmt.Printf("part.type=%v\n", part["type"])
	fmt.Printf("part.reason=%v\n", part["reason"])
	fmt.Printf("part.snapshot=%v\n", part["snapshot"])
	fmt.Printf("part.cost=%v\n", part["cost"])

	tokens, _ := part["tokens"].(map[string]any)
	fmt.Printf("tokens.input=%v\n", tokens["input"])
	fmt.Printf("tokens.output=%v\n", tokens["output"])
	fmt.Printf("tokens.reasoning=%v\n", tokens["reasoning"])

	cache, _ := tokens["cache"].(map[string]any)
	fmt.Printf("tokens.cache.read=%v\n", cache["read"])
	fmt.Printf("tokens.cache.write=%v\n", cache["write"])
}
`
	return nil
}
```
