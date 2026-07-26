## Preconditions
- An AgentEvent with Type "message" and text content.

## Steps
1. Set `req.AgentEventJSON` to a marshaled AgentEvent with `Type:"message"`.
2. Call `formatEventLine`.
3. Verify the formatted output contains the message text.

```go
import (
    "encoding/json"
    "testing"

    types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    event := types.AgentEvent{
        Type: types.ActionMessage,
        Text: "Hello, this is a test message",
    }
    data, _ := json.Marshal(event)
    req.AgentEventJSON = string(data)
    return nil
}
```
