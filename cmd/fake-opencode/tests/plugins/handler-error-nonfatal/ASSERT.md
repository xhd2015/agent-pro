---
label: e2e
---

## Expected
- Exit code 0 (plugin errors are non-fatal).
- Stderr contains "plugin handler failed".

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stderr, "plugin handler failed")
}
```
