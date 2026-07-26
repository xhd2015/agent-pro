---
label: e2e
---

## Expected
- The `session.created` hook fires.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"session.created"`)
    assertContains(t, resp.HookLog, `"session_id":"sess_hook"`)
}
```

