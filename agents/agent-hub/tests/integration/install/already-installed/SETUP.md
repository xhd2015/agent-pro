## Preconditions
- The plugin file already exists at `.opencode/plugins/agent-hub.ts`.

## Steps
1. Create the plugin file first.
2. Run `agent-hub integration install opencode`.
3. Verify it overwrites and reports the status.

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
    if err := os.WriteFile(filepath.Join(pluginsDir, "agent-hub.ts"), []byte("old"), 0644); err != nil {
        return err
    }
    req.Args = []string{"integration", "opencode", "install"}
    return nil
}
```
