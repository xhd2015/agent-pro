## Preconditions
- No sessions exist.

## Steps
1. Run `agent-hub sessions` on empty store.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    _ = t.Name()
    return nil
}
```
