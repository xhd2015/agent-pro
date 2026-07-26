---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains an opencode event with `"done":true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"sessionID":"sess_done"`)
    assertContains(t, resp.Stdout, `"done":true`)
}
```
