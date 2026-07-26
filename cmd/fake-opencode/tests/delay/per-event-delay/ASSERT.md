---
label: e2e
---

## Expected
- Exit code 0.
- Stdout contains the delayed message.
- Elapsed wall time >= 1500ms (verified in the overridden Run).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"text":"delayed"`)
}
```
