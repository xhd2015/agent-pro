## Steps
1. Run `agent-hub hook --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"hook", "--help"}
    return nil
}
```
