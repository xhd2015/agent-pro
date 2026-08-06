## Expected

- No error.
- Exactly 1 session: multiSessionID(0) with prompt text `inside`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListIDsNewestFirst(t, resp.List, []string{multiSessionID(0)})
	assertPromptCount(t, &resp.List[0], 1)
	assertPromptText(t, &resp.List[0], 0, "inside")
}
```
