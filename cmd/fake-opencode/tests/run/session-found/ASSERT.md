---
label: e2e
---

## Expected
- The command exits with code 0.
- The emitted event uses the session flag.
- No error is produced.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"sessionID":"sess_known"`)
}
```
