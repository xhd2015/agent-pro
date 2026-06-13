## Steps
1. Run `agent-hub integration disable opencode` when no plugin file exists.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "disable", "opencode"}
    return nil
}
```
