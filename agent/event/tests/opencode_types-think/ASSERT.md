## Expected
- One opencode event: type `reasoning` with session ID and think text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"sessionID":"sess_001"`)
	assertContains(t, resp.Output, `"thinking about the problem"`)
	assertContains(t, resp.Output, `"id":"evt_1"`)
}
```
