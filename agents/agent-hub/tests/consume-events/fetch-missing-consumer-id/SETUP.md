## Preconditions
- --consumer-id is omitted.

## Steps
1. Run `agent-hub fetch` without --consumer-id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"fetch"}
    return nil
}
```
