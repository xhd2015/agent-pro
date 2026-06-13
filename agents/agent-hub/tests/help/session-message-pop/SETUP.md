## Steps
1. Run `agent-hub session message pop --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "message", "pop", "--help"}
    return nil
}
```
