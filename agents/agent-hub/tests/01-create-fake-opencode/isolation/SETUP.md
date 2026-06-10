## Preconditions
- This branch tests host configuration isolation.

## Steps
1. Mark the test mode as isolation.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=isolation")
    return nil
}
```

