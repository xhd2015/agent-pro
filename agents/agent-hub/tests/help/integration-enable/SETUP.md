## Steps
1. Run `agent-hub integration enable --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "enable", "--help"}
    return nil
}
```
