---
label: e2e
---

## Expected
- The command succeeds (stdout_events is still functional).
- stdout contains the event output.
- stderr contains a deprecation warning about `stdout_events`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"deprecated still works"`)
    assertContains(t, resp.Stderr, "stdout_events")
    assertContains(t, resp.Stderr, "deprecat")
}
```
