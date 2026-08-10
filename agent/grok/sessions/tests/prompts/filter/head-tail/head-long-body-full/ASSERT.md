## Expected

- No error.
- Structured: 1 kept prompt; OmittedAfter=1.
- Formatted first body is **full** 220 x-run (no soft-cap `…` on body).
- Trailing omission marker `(...1 omitted...)`.
- Trailing newline.

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertPromptCount(t, resp.Single, 1)
	if resp.Single.OmittedAfter != 1 {
		t.Fatalf("OmittedAfter=%d want 1", resp.Single.OmittedAfter)
	}

	out := resp.Output
	assertTrailingNewline(t, out)
	xs := longPromptRunes(220)
	if !strings.Contains(out, xs) {
		t.Fatalf("head-kept long body must be full 220 x runes:\n%s", out)
	}
	if strings.Contains(out, "…") {
		t.Fatalf("default body must not soft-cap with … when MaxBody unset:\n%s", out)
	}
	assertContains(t, out, omissionMarker(1))
}
```
