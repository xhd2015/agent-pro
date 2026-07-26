## Steps
1. Set HOME to a temp directory.
2. Run `agent-hub integration install opencode --global`.
3. Verify plugin file was created under `$HOME/.config/opencode/plugins/agent-hub.ts`.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    fakeHome := filepath.Join(req.TempDir, "fake-home")
    if err := os.MkdirAll(fakeHome, 0755); err != nil {
        return err
    }
    req.Args = []string{"integration", "opencode", "install", "--global"}
    req.Env = append(req.Env, "HOME="+fakeHome)
    return nil
}
```
