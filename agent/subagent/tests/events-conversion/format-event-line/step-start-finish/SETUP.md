## Preconditions
- AgentEvents with Type "step_start" and "step_finish".

## Steps
1. Set `req.AgentEventJSON` to a marshaled AgentEvent with `Type:"step_start"`.
2. Call `formatEventLine`.
3. Verify the formatted output is non-empty (indicates the step event is rendered).

```go
import (
    "encoding/json"
    "testing"

    types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    event := types.AgentEvent{
        Type: types.ActionStepStart,
        Text: "Starting step",
    }
    data, _ := json.Marshal(event)
    req.AgentEventJSON = string(data)
    return nil
}
```
