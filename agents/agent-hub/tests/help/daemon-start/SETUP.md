## Steps
1. Run `agent-hub daemon start --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"daemon", "start", "--help"}
    return nil
}
```
