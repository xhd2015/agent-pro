## Expected
- JSON parses into Event correctly.
- All step_start part fields are populated with correct values.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "type=step_start")
	assertContains(t, resp.Output, "sessionID=sess_ss")
	assertContains(t, resp.Output, "timestamp=1718200000123")
	assertContains(t, resp.Output, "part.id=p1")
	assertContains(t, resp.Output, "part.sessionID=sess_ss")
	assertContains(t, resp.Output, "part.messageID=msg_1")
	assertContains(t, resp.Output, "part.type=step-start")
	assertContains(t, resp.Output, "part.snapshot=snap_abc")
}
```
