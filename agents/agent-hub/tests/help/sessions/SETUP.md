## Steps
1. Run `agent-hub sessions --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"sessions", "--help"}
    return nil
}
```
