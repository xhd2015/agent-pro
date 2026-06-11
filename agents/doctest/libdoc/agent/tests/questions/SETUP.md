## Preconditions
- Sub-agent may yield pending questions during implementation.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "TEST_GROUP=questions")
    return nil
}
```
