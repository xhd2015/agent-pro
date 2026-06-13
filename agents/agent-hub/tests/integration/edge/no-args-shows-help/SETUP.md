## Steps
1. Run `agent-hub integration` with no subcommand (no args).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration"}
    return nil
}
```
