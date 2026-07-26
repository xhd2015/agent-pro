## Expected
- One crush event: type `run_complete`.
- Session and message IDs present, no error, not cancelled.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"run_complete"`)
	assertContains(t, resp.Output, `"session_id":"sess_crush"`)
	assertNotContains(t, resp.Output, `"error"`)
}
```
