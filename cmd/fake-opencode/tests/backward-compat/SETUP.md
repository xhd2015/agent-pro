## Preconditions
- The mock config uses the legacy opencode-specific event format.
- fake-opencode must still accept and process this format correctly.

## Steps
1. Mark the test mode.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=backward-compat")
    return nil
}
```
