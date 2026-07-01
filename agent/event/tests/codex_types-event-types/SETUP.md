# Scenario

**Feature**: The `codex_types` package defines `Event`, `EventItem`, `EventType`, `ItemType` and all their constants

## Preconditions
- The `codex_types` package defines `Event`, `EventItem`, `EventType`, `ItemType` and all their constants.

## Steps
1. Print all constant values and marshal an `Event` with an `EventItem`.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	codex_types "github.com/xhd2015/agent-pro/agent/event/codex_types"
)

func Setup(t *testing.T, req *Request) error {
	var sb strings.Builder

	fmt.Fprintf(&sb, "EventStarted=%s\n", codex_types.EventStarted)
	fmt.Fprintf(&sb, "EventUpdated=%s\n", codex_types.EventUpdated)
	fmt.Fprintf(&sb, "EventCompleted=%s\n", codex_types.EventCompleted)
	fmt.Fprintf(&sb, "EventError=%s\n", codex_types.EventError)
	fmt.Fprintf(&sb, "ItemReasoning=%s\n", codex_types.ItemReasoning)
	fmt.Fprintf(&sb, "ItemCommandExecution=%s\n", codex_types.ItemCommandExecution)
	fmt.Fprintf(&sb, "ItemFileChange=%s\n", codex_types.ItemFileChange)
	fmt.Fprintf(&sb, "ItemMessage=%s\n", codex_types.ItemMessage)

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
	fmt.Fprintln(&sb, string(data))

	req.Output = sb.String()
	return nil
}
```
