## Preconditions
- An AgentEvent with a Timestamp field set.

## Steps
1. Set `req.AgentEventJSON` to a marshaled AgentEvent with `Timestamp:1718444400000`.
2. Call `formatEventLine`.
3. Verify the formatted output is non-empty.

```go
import (
    "encoding/json"
    "testing"

    types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    event := types.AgentEvent{
        Type:      types.ActionMessage,
        Text:      "Timestamped message",
        Timestamp: 1718444400000,
    }
    data, _ := json.Marshal(event)
    req.AgentEventJSON = string(data)
    return nil
}
```
