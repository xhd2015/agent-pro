## Preconditions
- The test uses `fake-codex exec --json --mock-config`.

## Steps
1. Configure the leaf-specific mock JSON file.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=mock-config")
    return nil
}
```
