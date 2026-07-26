## Steps
1. Run `agent-hub integration install unknown-runner`.
2. Verify error for unsupported runner.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    req.Args = []string{"integration", "unknown-runner", "install"}
    return nil
}
```
