## Preconditions
- `formatEventLine` renders AgentEvent JSON as human-readable strings.

## Steps
1. Set `req.Operation = "format_event"`.
2. Each leaf sets `req.AgentEventJSON` to a marshaled AgentEvent.
3. `Run` calls `subagent.TestExported_formatEventLine`, returns the formatted string.

```go
import (
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    req.Operation = "format_event"
    return nil
}
```
