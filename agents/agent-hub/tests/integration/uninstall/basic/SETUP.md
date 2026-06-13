## Steps
1. Pre-create `.opencode/plugins/agent-hub.ts`.
2. Run `agent-hub integration uninstall opencode`.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    pluginsDir := filepath.Join(req.TempDir, ".opencode", "plugins")
    if err := os.MkdirAll(pluginsDir, 0755); err != nil {
        return err
    }
    if err := os.WriteFile(filepath.Join(pluginsDir, "agent-hub.ts"), []byte("plugin"), 0644); err != nil {
        return err
    }
    req.Args = []string{"integration", "opencode", "uninstall"}
    return nil
}
```
