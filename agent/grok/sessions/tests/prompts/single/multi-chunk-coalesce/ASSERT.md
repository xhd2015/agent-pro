## Expected

- No error.
- Exactly **one** user prompt.
- Text is the concatenation of both chunks: `hello there`.
- Timestamp equals **first** chunk time (fixedNow − 20m), not the second.

## Errors

- None.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertPromptCount(t, resp.Single, 1)
	assertPromptText(t, resp.Single, 0, "hello there")
	assertPromptTimeUTC(t, resp.Single, 0, atFixed(-20*time.Minute))
}
```
