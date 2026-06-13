## Steps
1. Run `agent-hub integration install --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"integration", "install", "--help"}
    return nil
}
```
