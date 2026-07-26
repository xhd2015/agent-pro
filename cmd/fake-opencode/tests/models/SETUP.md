## Preconditions
- This branch tests the `models` command.

## Steps
1. Mark the test mode as models.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=models")
    return nil
}
```

