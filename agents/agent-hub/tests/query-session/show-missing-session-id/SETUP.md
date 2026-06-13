## Preconditions
- --session-id is omitted.

## Steps
1. Run `agent-hub session show --runner fake-opencode` (no --session-id).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "show", "--runner", "fake-opencode"}
    return nil
}
```
