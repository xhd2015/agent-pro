## Preconditions
- The test uses legacy `--script`.

## Steps
1. Configure a legacy fake Codex script.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=legacy-script")
    return nil
}
```
