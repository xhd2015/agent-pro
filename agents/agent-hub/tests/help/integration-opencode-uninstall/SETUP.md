## Steps
1. Run `agent-hub integration opencode uninstall --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "opencode", "uninstall", "--help"}
    return nil
}
```
