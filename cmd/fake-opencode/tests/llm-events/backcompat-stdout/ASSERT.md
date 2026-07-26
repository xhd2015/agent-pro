---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains the converted event.
- stderr contains deprecation warning (stdout_events is deprecated).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"backward compat works"`)
    assertContains(t, resp.Stderr, "deprecat")
}
```
