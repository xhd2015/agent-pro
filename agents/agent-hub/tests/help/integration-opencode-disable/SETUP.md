## Steps
1. Run `agent-hub integration opencode disable --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "opencode", "disable", "--help"}
    return nil
}
```
