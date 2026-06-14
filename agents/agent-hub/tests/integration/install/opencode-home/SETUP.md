## Steps
1. Set up a custom opencode base directory.
2. Run `agent-hub integration opencode install --opencode-home <dir>`.

```go
import (
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    opencodeHome := filepath.Join(req.TempDir, "fake-opencode-home")
    req.Args = []string{"integration", "opencode", "install", "--opencode-home", opencodeHome}
    return nil
}
```
