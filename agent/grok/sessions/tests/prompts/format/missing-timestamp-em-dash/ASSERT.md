## Expected

- Output contains a line with `[—]` (em dash U+2014) and text `untimed`.
- Trailing newline.
- Does **not** use a zero-calendar timestamp like `0001-01-01`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertTrailingNewline(t, resp.Output)
	assertContains(t, resp.Output, "[—]")
	assertContains(t, resp.Output, "untimed")
	assertNotContains(t, resp.Output, "0001-01-01")
	// Ensure em-dash form is the timestamp slot for the prompt line.
	found := false
	for _, ln := range strings.Split(resp.Output, "\n") {
		if strings.HasPrefix(ln, "[—]") && strings.Contains(ln, "untimed") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected line starting with [—] and containing untimed:\n%s", resp.Output)
	}
}
```
