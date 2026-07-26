## Steps
1. Pre-create `agent-hub.ts` in global plugins dir.
2. Run `agent-hub integration opencode disable --global`.

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
    pluginsDir := filepath.Join(fakeHome, ".config", "opencode", "plugins")
    if err := os.MkdirAll(pluginsDir, 0755); err != nil {
        return err
    }
    if err := os.WriteFile(filepath.Join(pluginsDir, "agent-hub.ts"), []byte("plugin"), 0644); err != nil {
        return err
    }
    req.Args = []string{"integration", "opencode", "disable", "--global"}
    req.Env = append(req.Env, "HOME="+fakeHome)
    return nil
}
```
