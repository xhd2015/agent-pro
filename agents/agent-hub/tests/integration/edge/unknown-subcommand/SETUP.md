## Steps
1. Run `agent-hub integration unknown-subcommand`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "unknown-subcommand"}
    return nil
}
```
