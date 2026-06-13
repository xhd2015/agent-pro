## Steps
1. Run `agent-hub notify --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"notify", "--help"}
    return nil
}
```
