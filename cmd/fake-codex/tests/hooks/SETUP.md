## Preconditions
- The test uses a hook recorder command from the mock config.

## Steps
1. Configure one or more hook events for the leaf timing.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=hooks")
    return nil
}
```
