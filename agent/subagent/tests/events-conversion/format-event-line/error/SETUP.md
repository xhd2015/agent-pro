## Preconditions
- An AgentEvent with Type "error" and error text.

## Steps
1. Set `req.AgentEventJSON` to a marshaled AgentEvent with `Type:"error"`.
2. Call `formatEventLine`.
3. Verify the formatted output contains the error text.

```go
import (
    "encoding/json"
    "testing"

    types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, req *Request) error {
    event := types.AgentEvent{
        Type: types.ActionError,
        Text: "connection refused",
    }
    data, _ := json.Marshal(event)
    req.AgentEventJSON = string(data)
    return nil
}
```
