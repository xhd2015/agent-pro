## Steps
1. Run `agent-hub integration disable --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "disable", "--help"}
    return nil
}
```
