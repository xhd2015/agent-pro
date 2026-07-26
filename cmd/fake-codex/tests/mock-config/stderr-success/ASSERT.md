---
label: e2e
---

## Expected
- The command succeeds.
- stderr contains the configured diagnostic.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stderr, "diagnostic line")
}
```

