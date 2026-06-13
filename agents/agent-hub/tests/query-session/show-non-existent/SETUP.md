## Preconditions
- No session exists.

## Steps
1. Run session show for non-existent ID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "show", "--runner", "fake-opencode", "--session-id", "nosuch"}
    return nil
}
```
