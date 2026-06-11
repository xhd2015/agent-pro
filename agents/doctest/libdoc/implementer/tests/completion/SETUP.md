## Preconditions
- Sub-agent completes its work successfully.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "TEST_GROUP=completion")
    return nil
}
```
