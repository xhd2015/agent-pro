## Steps
1. Pre-create `.opencode/plugins/agent-hub.ts`.
2. Run `agent-hub integration disable opencode`.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    pluginsDir := filepath.Join(req.TempDir, ".opencode", "plugins")
    if err := os.MkdirAll(pluginsDir, 0755); err != nil {
        return err
    }
    if err := os.WriteFile(filepath.Join(pluginsDir, "agent-hub.ts"), []byte("plugin"), 0644); err != nil {
        return err
    }
    req.Args = []string{"integration", "opencode", "disable"}
    return nil
}
```
