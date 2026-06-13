## Preconditions
- hook notify is called with a known runner but unknown event.

## Steps
1. Run `agent-hub hook notify --runner opencode --event BogusEvent` with stdin payload.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"hook", "notify", "--runner", "opencode", "--event", "BogusEvent"}
    req.Stdin = `{"session_id":"s1"}`
    return nil
}
```
