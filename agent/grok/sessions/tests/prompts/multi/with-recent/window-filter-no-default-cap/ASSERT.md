## Expected

- No error.
- Exactly **12** sessions (all in-window); multiSessionID(90) excluded.
- Not clipped to default 10.
- Newest-first order multiSessionID(0)..(11).
- Each returned session has exactly 1 user prompt with text `in-window`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	want := make([]string, 12)
	for i := 0; i < 12; i++ {
		want[i] = multiSessionID(i)
	}
	assertListIDsNewestFirst(t, resp.List, want)
	for i := range resp.List {
		sp := &resp.List[i]
		assertPromptCount(t, sp, 1)
		assertPromptText(t, sp, 0, "in-window")
	}
}
```
