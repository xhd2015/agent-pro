## Preconditions
- An AgentEvent with Type "think" and reasoning text.

## Steps
1. Set `req.AgentEventJSON` to a marshaled AgentEvent with `Type:"think"`.
2. Call `formatEventLine`.
3. Verify the formatted output contains the reasoning text.

```go
import (
    "encoding/json"
    "testing"

    types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    event := types.AgentEvent{
        Type: types.ActionThink,
        Text: "Let me think about this problem carefully...",
    }
    data, _ := json.Marshal(event)
    req.AgentEventJSON = string(data)
    return nil
}
```
