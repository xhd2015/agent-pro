# Scenario

**Feature**: coalesce consecutive `ActionMessage` deltas into one assistant block

## Preconditions
- Control case: consecutive `ActionMessage` streaming deltas are already coalesced by `FormatState`.
- Same harness as grok-thought-deltas; verifies message coalescing works.

## Steps
1. Set `req.Lines` to two consecutive `ActionMessage` deltas: "Hel", "lo".

```go
import (
	"encoding/json"
	"testing"

	eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
	for _, text := range []string{"Hel", "lo"} {
		data, _ := json.Marshal(eventtypes.AgentEvent{
			Type: eventtypes.ActionMessage,
			Text: text,
		})
		req.Lines = append(req.Lines, string(data))
	}
	return nil
}
```