## Steps
1. Run `agent-hub --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"--help"}
    return nil
}
```
