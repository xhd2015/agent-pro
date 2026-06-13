## Steps
1. Run `agent-hub session message list --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "message", "list", "--help"}
    return nil
}
```
