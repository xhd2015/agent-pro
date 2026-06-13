## Steps
1. Run `agent-hub partitions --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Args = []string{"partitions", "--help"}
    return nil
}
```
