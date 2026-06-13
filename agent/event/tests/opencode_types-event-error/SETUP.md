## Preconditions
- The `opencode_types` package defines `Event` and `ErrorDetail` structs.

## Steps
1. Create an `Event` with type `error` and an `ErrorDetail` containing a message.
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
		Type:      "error",
		SessionID: "sess_e1",
		Error: &opencode_types.ErrorDetail{
			Name: "Error",
			Data: &opencode_types.ErrorData{
				Message: "something went wrong",
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
