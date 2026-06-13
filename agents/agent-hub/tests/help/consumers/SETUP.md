## Steps
1. Run `agent-hub consumers --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"consumers", "--help"}
    return nil
}
```
