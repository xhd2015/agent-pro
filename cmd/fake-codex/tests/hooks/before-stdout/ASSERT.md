---
label: e2e
---

## Expected
- The `before_stdout` hook fires.
- stdout is still emitted.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"BeforeStdout"`)
    assertContains(t, resp.Stdout, `"after hook"`)
}
```

