## Expected

- No error.
- List length 0 (or only empty survivors filtered before format).
- Output contains `No user prompts found`.
- Trailing newline.
- Does not contain the original prompt text `hello world`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListLen(t, resp.List, 0)
	assertContains(t, resp.Output, "No user prompts found")
	assertTrailingNewline(t, resp.Output)
	assertNotContains(t, resp.Output, "hello world")
}
```
