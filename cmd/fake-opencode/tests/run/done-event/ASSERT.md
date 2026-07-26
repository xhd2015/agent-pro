---
label: e2e
---

## Expected
- The command succeeds.
- The done event is emitted with session ID and `"done":true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"done":true`)
	assertContains(t, resp.Stdout, `"sessionID":"sess_done"`)
}
```
