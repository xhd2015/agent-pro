## Steps
1. Run `agent-hub integration opencode status --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "opencode", "status", "--help"}
    return nil
}
```
