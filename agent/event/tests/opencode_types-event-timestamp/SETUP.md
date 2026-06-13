## Preconditions
- The `opencode_types` package defines `Event` with a `Timestamp` field.

## Steps
1. Parse a raw JSON string containing a timestamp (matching real opencode output envelope).
2. Marshal back and verify timestamp is preserved.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	"strings"
	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func main() {
	raw := ` + "`" + `{"type":"text","timestamp":1718200000999,"sessionID":"sess_ts","part":{"id":"p3","type":"text","text":"hello with timestamp"}}` + "`" + `
	var evt opencode_types.Event
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		fmt.Println("PARSE ERROR:", err)
		return
	}
	fmt.Printf("type=%s\n", evt.Type)
	fmt.Printf("sessionID=%s\n", evt.SessionID)
	fmt.Printf("timestamp=%d\n", evt.Timestamp)

	// re-marshal and verify timestamp is emitted
	data, _ := json.Marshal(evt)
	jsonStr := string(data)
	fmt.Printf("has_timestamp=%v\n", strings.Contains(jsonStr, ` + "`" + `"timestamp"` + "`" + `))
}
`
	return nil
}
```
