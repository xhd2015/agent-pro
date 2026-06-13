## Expected
- JSON parses into Event correctly.
- All step_start part fields are populated with correct values.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "type=step_start")
	assertContains(t, resp.Stdout, "sessionID=sess_ss")
	assertContains(t, resp.Stdout, "timestamp=1718200000123")
	assertContains(t, resp.Stdout, "part.id=p1")
	assertContains(t, resp.Stdout, "part.sessionID=sess_ss")
	assertContains(t, resp.Stdout, "part.messageID=msg_1")
	assertContains(t, resp.Stdout, "part.type=step-start")
	assertContains(t, resp.Stdout, "part.snapshot=snap_abc")
}
```
