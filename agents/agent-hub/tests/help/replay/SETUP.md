## Steps
1. Run `agent-hub replay --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"replay", "--help"}
    return nil
}
```
