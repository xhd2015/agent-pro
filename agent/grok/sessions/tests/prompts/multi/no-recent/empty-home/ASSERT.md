## Expected

- No error.
- List length 0.
- Output contains `No user prompts found`.
- Output ends with trailing newline.
- No emoji USER chrome.

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
	assertNotContains(t, resp.Output, "👤")
}
```
