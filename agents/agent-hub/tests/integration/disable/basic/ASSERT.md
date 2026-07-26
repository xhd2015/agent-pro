## Expected
- Exit code 0.
- Stdout indicates disabled.
- File `agent-hub.ts` renamed to `agent-hub.ts.disabled`.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Disabled")

    pluginsDir := filepath.Join(req.TempDir, ".opencode", "plugins")
    if _, err := os.Stat(filepath.Join(pluginsDir, "agent-hub.ts.disabled")); os.IsNotExist(err) {
        t.Fatal("agent-hub.ts.disabled should exist after disable")
    }
    if _, err := os.Stat(filepath.Join(pluginsDir, "agent-hub.ts")); err == nil {
        t.Fatal("agent-hub.ts should not exist after disable")
    }
}
```
