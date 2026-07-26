---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains an opencode reasoning event with session ID and text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"reasoning"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_basic"`)
    assertContains(t, resp.Stdout, `"thinking about it"`)
}
```
