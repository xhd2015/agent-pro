## Expected
- Exit code 0.
- Stdout indicates enabled.
- File `agent-hub.ts.disabled` renamed to `agent-hub.ts`.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Enabled")

    pluginsDir := filepath.Join(req.TempDir, ".opencode", "plugins")
    if _, err := os.Stat(filepath.Join(pluginsDir, "agent-hub.ts")); os.IsNotExist(err) {
        t.Fatal("agent-hub.ts should exist after enable")
    }
    if _, err := os.Stat(filepath.Join(pluginsDir, "agent-hub.ts.disabled")); err == nil {
        t.Fatal("agent-hub.ts.disabled should not exist after enable")
    }
}
```
