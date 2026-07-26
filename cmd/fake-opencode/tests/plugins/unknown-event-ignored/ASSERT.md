---
label: e2e
---

## Expected
- Exit code 0.
- No errors in stderr about unknown events.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
}
```
