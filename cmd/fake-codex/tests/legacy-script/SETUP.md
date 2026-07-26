## Preconditions
- The test uses legacy `--script`.

## Steps
1. Configure a legacy fake Codex script.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=legacy-script")
    return nil
}
```
