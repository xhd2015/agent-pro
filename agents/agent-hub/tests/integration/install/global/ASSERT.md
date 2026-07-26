## Expected
- Exit code 0.
- Stdout indicates successful installation.
- File `$HOME/.config/opencode/plugins/agent-hub.ts` exists.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Installed:")
    assertContains(t, resp.Stdout, "agent-hub.ts")

    fakeHome := filepath.Join(req.TempDir, "fake-home")
    pluginPath := filepath.Join(fakeHome, ".config", "opencode", "plugins", "agent-hub.ts")
    if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
        t.Fatalf("plugin not created at %s", pluginPath)
    }
}
```
