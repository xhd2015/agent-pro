## Expected
- Exit code 0.
- Stdout indicates successful uninstallation.
- File `.opencode/plugins/agent-hub.ts` no longer exists.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Uninstalled")

    pluginPath := filepath.Join(req.TempDir, ".opencode", "plugins", "agent-hub.ts")
    if _, err := os.Stat(pluginPath); err == nil {
        t.Fatalf("plugin still exists at %s", pluginPath)
    }
}
```
