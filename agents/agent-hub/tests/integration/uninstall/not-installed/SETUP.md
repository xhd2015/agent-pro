## Steps
1. Run `agent-hub integration uninstall opencode` when no plugin file exists.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "uninstall", "opencode"}
    return nil
}
```
