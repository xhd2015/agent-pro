## Preconditions
- hook notify is called with an unknown runner.

## Steps
1. Run `agent-hub hook notify --runner unknown --event SessionStart` with stdin payload.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"hook", "notify", "--runner", "unknown", "--event", "SessionStart"}
    req.Stdin = `{"session_id":"s1"}`
    return nil
}
```
