## Expected

- No error.
- Exactly 3 sessions: multiSessionID(0), (1), (2) in that order.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListIDsNewestFirst(t, resp.List, []string{
		multiSessionID(0),
		multiSessionID(1),
		multiSessionID(2),
	})
}
```
