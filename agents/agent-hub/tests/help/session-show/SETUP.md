## Steps
1. Run `agent-hub session show --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "show", "--help"}
    return nil
}
```
