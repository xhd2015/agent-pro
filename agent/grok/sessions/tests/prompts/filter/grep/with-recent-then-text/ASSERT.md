## Expected

- No error.
- Exactly 1 session.
- Exactly 1 prompt: `hit-new`.
- `hit-old` excluded by recent window (even though it matches grep).
- `miss-new` excluded by grep.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListLen(t, resp.List, 1)
	sp := &resp.List[0]
	assertPromptCount(t, sp, 1)
	assertPromptText(t, sp, 0, "hit-new")
}
```
