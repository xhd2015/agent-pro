## Steps
1. Run `agent-hub fetch --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"fetch", "--help"}
    return nil
}
```
