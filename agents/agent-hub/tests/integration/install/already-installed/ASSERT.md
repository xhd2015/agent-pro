## Expected
- Exit code 0.
- Stdout indicates plugin was already installed and overwritten.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Installed")

    pluginPath := filepath.Join(req.TempDir, ".opencode", "plugins", "agent-hub.ts")
    if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
        t.Fatalf("plugin not found at %s", pluginPath)
    }
}
```
