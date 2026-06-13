## Steps
1. Run `agent-hub integration status` (no runner arg) — lists all supported runners.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "status"}
    return nil
}
```
