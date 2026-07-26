## Expected
- JSON contains `"type":"reasoning"`, `"sessionID":"sess_r1"`, and a `part` with reasoning text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"sessionID":"sess_r1"`)
	assertContains(t, resp.Output, `"id":"evt_r1"`)
	assertContains(t, resp.Output, `"thinking step by step"`)
}
```
