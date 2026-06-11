## Preconditions
- This branch tests hook firing from mock config.

## Steps
1. Mark the test mode as hooks.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=hooks")
    return nil
}
```

