## Preconditions
- Sub-agent encounters an error and exits non-zero.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "TEST_GROUP=failure")
    return nil
}
```
