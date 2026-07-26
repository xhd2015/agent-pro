---
label: e2e
---

## Expected
- The text event is emitted as JSONL with session ID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"text"`)
    assertContains(t, resp.Stdout, `"sessionID":"sess_text"`)
    assertContains(t, resp.Stdout, `"fake opencode answered"`)
}
```

