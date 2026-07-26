## Preconditions
- The mock config uses the legacy codex-specific event format.
- fake-codex must still accept and process this format correctly.

## Steps
1. Mark the test mode.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=backward-compat")
    return nil
}
```
