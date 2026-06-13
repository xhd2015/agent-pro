## Steps
1. Pre-create `.opencode/plugins/agent-hub.ts.disabled`.
2. Run `agent-hub integration enable opencode`.

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
    if err := os.WriteFile(filepath.Join(pluginsDir, "agent-hub.ts.disabled"), []byte("plugin"), 0644); err != nil {
        return err
    }
    req.Args = []string{"integration", "enable", "opencode"}
    return nil
}
```
