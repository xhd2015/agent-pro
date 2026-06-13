## Steps
1. Run `agent-hub daemon --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"daemon", "--help"}
    return nil
}
```
