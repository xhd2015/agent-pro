## Expected
- Exit code 0.
- Stdout indicates disabled.
- Global `.ts` → `.ts.disabled`.

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Disabled")

    fakeHome := filepath.Join(req.TempDir, "fake-home")
    pluginsDir := filepath.Join(fakeHome, ".config", "opencode", "plugins")
    if _, err := os.Stat(filepath.Join(pluginsDir, "agent-hub.ts.disabled")); os.IsNotExist(err) {
        t.Fatal("agent-hub.ts.disabled should exist after disable")
    }
    if _, err := os.Stat(filepath.Join(pluginsDir, "agent-hub.ts")); err == nil {
        t.Fatal("agent-hub.ts should not exist after disable")
    }
}
```
