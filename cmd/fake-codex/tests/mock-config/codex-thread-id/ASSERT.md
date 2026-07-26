---
label: e2e
---

## Expected
- Exit code 0.
- The marker file contains `sess_thread_test` (the session_id set as CODEX_THREAD_ID).

```go
import (
    "os"
    "path/filepath"
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)

    marker := filepath.Join(req.TempDir, "thread-id.txt")
    data, readErr := os.ReadFile(marker)
    if readErr != nil {
        t.Fatalf("read marker file: %v", readErr)
    }
    got := strings.TrimSpace(string(data))
    if got != "sess_thread_test" {
        t.Fatalf("CODEX_THREAD_ID = %q, want %q", got, "sess_thread_test")
    }
}
```
