## Expected

- No error.
- Exactly 2 prompts: `keep-this`, `also-keep`.
- `drop-noise` absent.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertPromptCount(t, resp.Single, 2)
	assertPromptText(t, resp.Single, 0, "keep-this")
	assertPromptText(t, resp.Single, 1, "also-keep")
}
```
