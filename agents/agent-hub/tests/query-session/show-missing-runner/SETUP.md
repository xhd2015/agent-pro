## Preconditions
- --runner is omitted.

## Steps
1. Run `agent-hub session show --session-id s1` (no --runner).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "show", "--session-id", "s1"}
    return nil
}
```
