## Expected

- No error.
- Exactly **10** sessions returned.
- Newest-first: multiSessionID(0) … multiSessionID(9).
- Session multiSessionID(10)..(14) not present.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	want := make([]string, 10)
	for i := 0; i < 10; i++ {
		want[i] = multiSessionID(i)
	}
	assertListIDsNewestFirst(t, resp.List, want)
	assertSessionOrderByLastActive(t, resp.List)
}
```
