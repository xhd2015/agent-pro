## Steps
1. Run `agent-hub integration opencode install --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "opencode", "install", "--help"}
    return nil
}
```
