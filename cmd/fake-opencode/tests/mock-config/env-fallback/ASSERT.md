---
label: e2e
---

## Expected
- The env fallback mock config is used.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"from env"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_env"`)
}
```

