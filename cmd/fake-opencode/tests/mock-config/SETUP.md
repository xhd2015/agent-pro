## Preconditions
- This branch tests mock config resolution.

## Steps
1. Mark the test mode as mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=mock-config")
    return nil
}
```

