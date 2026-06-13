## Steps
1. Run `agent-hub integration enable opencode` when no plugin file exists.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "opencode", "enable"}
    return nil
}
```
