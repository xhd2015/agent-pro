---
label: e2e
---

## Expected
- The `after_stdout` hook fires.
- stdout is emitted.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"before hook"`)
    assertContains(t, resp.HookLog, `"event":"AfterStdout"`)
}
```

