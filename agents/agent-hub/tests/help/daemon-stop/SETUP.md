## Steps
1. Run `agent-hub daemon stop --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"daemon", "stop", "--help"}
    return nil
}
```
