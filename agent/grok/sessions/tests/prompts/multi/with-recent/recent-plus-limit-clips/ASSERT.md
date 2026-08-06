## Expected

- No error.
- Exactly **2** sessions: multiSessionID(0), multiSessionID(1).

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
	})
}
```
