## Expected
- One opencode event: type `text` with session ID and message text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"here is the result"`)
	assertContains(t, resp.Output, `"id":"evt_2"`)
}
```
