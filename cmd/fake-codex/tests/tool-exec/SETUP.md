## Preconditions
- This branch tests real tool execution during `fake-codex exec --json --mock-config`.
- Uses the existing harness from the parent SETUP.md (Request, Response, Run, writeMockConfig, etc.).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=tool-exec")
    return nil
}
```
