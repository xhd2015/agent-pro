## Steps
1. Run `agent-hub integration status unknown-runner`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "unknown-runner", "status"}
    return nil
}
```
