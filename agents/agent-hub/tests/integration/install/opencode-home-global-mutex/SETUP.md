## Steps
1. Run `agent-hub integration opencode install --opencode-home <dir> --global`.
2. This combination must be rejected as mutually exclusive.

```go
import (
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    opencodeHome := filepath.Join(req.TempDir, "fake-opencode-home")
    req.Args = []string{"integration", "opencode", "install", "--opencode-home", opencodeHome, "--global"}
    return nil
}
```
