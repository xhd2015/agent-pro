---
label: e2e
---

## Expected
- The command succeeds.
- The tool_use event for write shows status completed.
- The file is actually created on disk.

```go
import (
    "os"
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"tool":"write"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
    outPath := filepath.Join(req.TempDir, "written.txt")
    if _, statErr := os.Stat(outPath); statErr != nil {
        t.Fatalf("written file not found: %v", statErr)
    }
}
```
