## Preconditions
- **Reproduces**: Grok `thought` streaming writes per-token `ActionThink` events to events.jsonl.
- Each delta is a single word. `FormatState` should coalesce them into one thinking block.

## Steps
1. Set `req.Lines` to six consecutive `ActionThink` events with word deltas.
2. Deltas simulate grok streaming: "The", " user", " wants", " me", " to", " act".

```go
import (
	"encoding/json"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	words := []string{"The", " user", " wants", " me", " to", " act"}
	for _, w := range words {
		data, _ := json.Marshal(eventtypes.AgentEvent{
			Type: eventtypes.ActionThink,
			Text: w,
		})
		req.Lines = append(req.Lines, string(data))
	}
	return nil
}
```