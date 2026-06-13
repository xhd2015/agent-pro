## Steps
1. Run `agent-hub status --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"status", "--help"}
    return nil
}
```
