## Steps
1. Run `agent-hub integration status opencode` with no plugin file.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "status", "opencode"}
    return nil
}
```
