## Expected
- Exit code 0.
- Stdout indicates successful uninstallation.
- Global plugin file removed.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Uninstalled")

    fakeHome := filepath.Join(req.TempDir, "fake-home")
    pluginPath := filepath.Join(fakeHome, ".config", "opencode", "plugins", "agent-hub.ts")
    if _, err := os.Stat(pluginPath); err == nil {
        t.Fatalf("plugin still exists at %s", pluginPath)
    }
}
```
