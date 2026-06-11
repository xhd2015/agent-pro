## Preconditions
- The test uses `fake-codex exec --json --mock-config`.

## Steps
1. Configure the leaf-specific mock JSON file.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=mock-config")
    return nil
}
```
