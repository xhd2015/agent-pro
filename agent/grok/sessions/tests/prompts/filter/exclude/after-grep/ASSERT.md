## Expected

- No error.
- Exactly 1 prompt: `foo-ok`.
- `foobar-bad` matched grep but dropped by exclude.
- `other` never matched grep.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertPromptCount(t, resp.Single, 1)
	assertPromptText(t, resp.Single, 0, "foo-ok")
}
```
