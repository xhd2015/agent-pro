## Expected
- Exit code 0.
- Stdout indicates successful installation.
- File `$HOME/.config/opencode/plugins/agent-hub.ts` exists.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Installed")

    fakeHome := filepath.Join(req.TempDir, "fake-home")
    pluginPath := filepath.Join(fakeHome, ".config", "opencode", "plugins", "agent-hub.ts")
    if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
        t.Fatalf("plugin not created at %s", pluginPath)
    }
}
```
