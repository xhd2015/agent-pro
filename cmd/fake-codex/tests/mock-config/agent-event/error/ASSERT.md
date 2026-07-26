---
label: e2e
---

## Expected
- The command exits with code 3.
- stdout contains a codex error event with the message in the `message` field (not raw `text`).
- stderr contains the scripted error string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode != 3 {
        t.Fatalf("exit code = %d, want 3; stderr=%s", resp.ExitCode, resp.Stderr)
    }
    assertContains(t, resp.Stdout, `"type":"error"`)
    assertContains(t, resp.Stdout, `"message":"execution failed"`)
    assertNotContains(t, resp.Stdout, `"type":"error","text":"execution failed"`)
    assertContains(t, resp.Stderr, "something went wrong")
}
```
