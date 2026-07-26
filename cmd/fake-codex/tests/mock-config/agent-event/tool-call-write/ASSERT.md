---
label: e2e
---

## Expected
- The command succeeds.
- The completed codex event contains a file_change item.

```go
import (
    "path/filepath"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"file_change"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
    outPath := filepath.Join(req.TempDir, "output.txt")
    assertContains(t, resp.Stdout, outPath)
}
```
