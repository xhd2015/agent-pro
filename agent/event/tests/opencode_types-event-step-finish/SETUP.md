# Scenario

**Feature**: The `opencode_types` package defines `Event`, `StepFinishPart`, `Tokens`, and `CacheTokens`

## Preconditions
- The `opencode_types` package defines `Event`, `StepFinishPart`, `Tokens`, and `CacheTokens`.

## Steps
1. Parse a raw JSON string (matching real opencode step_finish output) into `Event`.
2. Verify nested fields like reason, cost, and token breakdown are populated.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
)

func Setup(t *testing.T, req *Request) error {
	raw := `{"type":"step_finish","timestamp":1718200000456,"sessionID":"sess_sf","part":{"id":"p2","sessionID":"sess_sf","messageID":"msg_2","type":"step-finish","reason":"stop","snapshot":"snap_xyz","cost":0.015,"tokens":{"input":120,"output":80,"reasoning":40,"cache":{"read":10,"write":5}}}}`
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
	fmt.Fprintf(&sb, "part.type=%v\n", part["type"])
	fmt.Fprintf(&sb, "part.reason=%v\n", part["reason"])
	fmt.Fprintf(&sb, "part.snapshot=%v\n", part["snapshot"])
	fmt.Fprintf(&sb, "part.cost=%v\n", part["cost"])

	tokens, _ := part["tokens"].(map[string]any)
	fmt.Fprintf(&sb, "tokens.input=%v\n", tokens["input"])
	fmt.Fprintf(&sb, "tokens.output=%v\n", tokens["output"])
	fmt.Fprintf(&sb, "tokens.reasoning=%v\n", tokens["reasoning"])

	cache, _ := tokens["cache"].(map[string]any)
	fmt.Fprintf(&sb, "tokens.cache.read=%v\n", cache["read"])
	fmt.Fprintf(&sb, "tokens.cache.write=%v\n", cache["write"])

	req.Output = sb.String()
	return nil
}
```
