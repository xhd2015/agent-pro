---
label: e2e
---

## Expected
- The `before_exit` hook fires before process completion.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"BeforeExit"`)
}
```

