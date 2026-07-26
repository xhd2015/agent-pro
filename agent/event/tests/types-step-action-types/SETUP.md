# Scenario

**Feature**: The `types` package defines `ActionStepStart` and `ActionStepFinish`

## Preconditions
- The `types` package defines `ActionStepStart` and `ActionStepFinish`.

## Steps
1. Create `AgentEvent` values with `step_start` and `step_finish` action types.
2. Marshal to JSON and verify the type strings.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	var sb strings.Builder
	startEvt := types.AgentEvent{
		ID:        "evt_ss",
		Type:      types.ActionStepStart,
		Timestamp: 1718200000123,
	}
	startData, _ := json.Marshal(startEvt)
	fmt.Fprintf(&sb, "step_start=%s\n", string(startData))

	finishEvt := types.AgentEvent{
		ID:        "evt_sf",
		Type:      types.ActionStepFinish,
		Timestamp: 1718200000456,
	}
	finishData, _ := json.Marshal(finishEvt)
	fmt.Fprintf(&sb, "step_finish=%s\n", string(finishData))
	req.Output = sb.String()
	return nil
}
```
