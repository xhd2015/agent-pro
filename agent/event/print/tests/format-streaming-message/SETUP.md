## Preconditions
- **Reproduces**: AgentEvent ActionMessage with PhaseUpdate deltas produce separate formatted blocks.
- When pi streaming deltas are converted to AgentEvent format via `FromPi`, each delta becomes
  an AgentEvent with `Type=message`, `Phase=update`. When `FormatTraceLine` processes these
  via the AgentEvent primary path, `FormatAgentEvent` does NOT coalesce them, producing
  fractional output (one block per delta).

## Steps
1. Set `req.Line` to a JSONL AgentEvent with `Type:"message"`, `Phase:"update"`, and a small delta text.
2. Call `print.FormatTraceLine`.
3. Verify the formatted output exists but note it cannot coalesce with prior/following events.

```go
import (
    "encoding/json"
    "testing"

    eventtypes "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
    // Simulate an AgentEvent produced by FromPi from a streaming delta
    event := eventtypes.AgentEvent{
        Type:  eventtypes.ActionMessage,
        Phase: eventtypes.PhaseUpdate,
        Text:  " a small delta",
    }
    data, _ := json.Marshal(event)
    req.Line = string(data)
    return nil
}
```
