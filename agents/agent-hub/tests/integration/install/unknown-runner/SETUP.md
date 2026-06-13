## Steps
1. Run `agent-hub integration install unknown-runner`.
2. Verify error for unsupported runner.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "unknown-runner", "install"}
    return nil
}
```
