## Steps
1. Run `agent-hub session --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"session", "--help"}
    return nil
}
```
