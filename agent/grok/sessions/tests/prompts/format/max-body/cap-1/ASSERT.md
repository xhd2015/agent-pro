## Expected

- No error.
- Body content is exactly **1** rune (`a`) followed by `…`.
- Rest of original text (`bcdefghi`) not present.
- Trailing newline.

## Errors

- None.

```go
import (
	"strings"
	"testing"
	"unicode/utf8"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	out := resp.Output
	assertTrailingNewline(t, out)

	body := out
	if i := strings.Index(out, "] "); i >= 0 {
		body = strings.TrimSuffix(out[i+2:], "\n")
		if j := strings.Index(body, "\n"); j >= 0 {
			body = body[:j]
		}
	}
	if body != "a…" {
		t.Fatalf("MaxBody=1 want body %q, got %q (runes=%d)", "a…", body, utf8.RuneCountInString(body))
	}
	assertNotContains(t, out, "bcdefghi")
	assertNotContains(t, out, "abcdefghi")
}
```
