# Scenario

**Feature**: The `opencode_types` package defines `Event` with a `Timestamp` field

## Preconditions
- The `opencode_types` package defines `Event` with a `Timestamp` field.

## Steps
1. Parse a raw JSON string containing a timestamp (matching real opencode output envelope).
2. Marshal back and verify timestamp is preserved.

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	opencode_types "github.com/xhd2015/agent-pro/agent/event/opencode_types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	raw := `{"type":"text","timestamp":1718200000999,"sessionID":"sess_ts","part":{"id":"p3","type":"text","text":"hello with timestamp"}}`
	var evt opencode_types.Event
	if err := json.Unmarshal([]byte(raw), &evt); err != nil {
		return err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "type=%s\n", evt.Type)
	fmt.Fprintf(&sb, "sessionID=%s\n", evt.SessionID)
	fmt.Fprintf(&sb, "timestamp=%d\n", evt.Timestamp)

	data, _ := json.Marshal(evt)
	jsonStr := string(data)
	fmt.Fprintf(&sb, "has_timestamp=%v\n", strings.Contains(jsonStr, `"timestamp"`))

	req.Output = sb.String()
	return nil
}
```
