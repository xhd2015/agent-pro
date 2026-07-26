## Expected
- Exit code 0.
- Stdout indicates installation path under `<dir>/plugins/agent-hub.ts`.
- Plugin file exists at `<dir>/plugins/agent-hub.ts`.

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

    opencodeHome := filepath.Join(req.TempDir, "fake-opencode-home")
    pluginPath := filepath.Join(opencodeHome, "plugins", "agent-hub.ts")
    if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
        t.Fatalf("plugin not created at %s", pluginPath)
    }
}
```
