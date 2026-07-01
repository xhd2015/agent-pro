# Scenario

**Feature**: The `opencode_types` package defines `Event` and `StepStartPart`

## Preconditions
- The `opencode_types` package defines `Event` and `StepStartPart`.

## Steps
1. Parse a raw JSON string (matching real opencode step_start output) into `Event`.
2. Verify the step_start part fields are populated correctly.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func Setup(t *testing.T, req *Request) error {
	raw := `{"type":"step_start","timestamp":1718200000123,"sessionID":"sess_ss","part":{"id":"p1","sessionID":"sess_ss","messageID":"msg_1","type":"step-start","snapshot":"snap_abc"}}`
	var evt opencode_types.Event
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		return err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "type=%s\n", evt.Type)
	fmt.Fprintf(&sb, "sessionID=%s\n", evt.SessionID)
	fmt.Fprintf(&sb, "timestamp=%d\n", evt.Timestamp)

	part, ok := evt.Part.(map[string]any)
	if !ok {
		fmt.Fprint(&sb, "PART_TYPE=map\n")
	}
	fmt.Fprintf(&sb, "part.id=%v\n", part["id"])
	fmt.Fprintf(&sb, "part.sessionID=%v\n", part["sessionID"])
	fmt.Fprintf(&sb, "part.messageID=%v\n", part["messageID"])
	fmt.Fprintf(&sb, "part.type=%v\n", part["type"])
	fmt.Fprintf(&sb, "part.snapshot=%v\n", part["snapshot"])

	req.Output = sb.String()
	return nil
}
```
