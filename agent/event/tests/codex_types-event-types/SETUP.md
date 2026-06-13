## Preconditions
- The `codex_types` package defines `Event`, `EventItem`, `EventType`, `ItemType` and all their constants.

## Steps
1. Print all constant values and marshal an `Event` with an `EventItem`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.MainGo = `package main

import (
	"encoding/json"
	"fmt"
	codex_types "github.com/xhd2015/agent-pro/agent/event/codex_types"
)

func main() {
	fmt.Printf("EventStarted=%s\n", codex_types.EventStarted)
	fmt.Printf("EventUpdated=%s\n", codex_types.EventUpdated)
	fmt.Printf("EventCompleted=%s\n", codex_types.EventCompleted)
	fmt.Printf("EventError=%s\n", codex_types.EventError)
	fmt.Printf("ItemReasoning=%s\n", codex_types.ItemReasoning)
	fmt.Printf("ItemCommandExecution=%s\n", codex_types.ItemCommandExecution)
	fmt.Printf("ItemFileChange=%s\n", codex_types.ItemFileChange)
	fmt.Printf("ItemMessage=%s\n", codex_types.ItemMessage)

	ec := 0
	evt := codex_types.Event{
		Type: codex_types.EventCompleted,
		Item: &codex_types.EventItem{
			ID:               "item_1",
			Type:             codex_types.ItemCommandExecution,
			Command:          "go test",
			AggregatedOutput: "ok",
			ExitCode:         &ec,
			Status:           "completed",
		},
	}
	data, _ := json.Marshal(evt)
	fmt.Println(string(data))
}
`
	return nil
}
```
