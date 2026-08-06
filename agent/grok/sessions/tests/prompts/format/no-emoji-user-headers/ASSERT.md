## Expected

- No error.
- Output contains `plain prompt`.
- Output does **not** contain `👤`.
- Output does **not** contain a multi-line USER card marker like
  `──── USER` or `Role: user` blocks typical of full session view.
- Trailing newline.

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
	assertTrailingNewline(t, resp.Output)
	assertContains(t, resp.Output, "plain prompt")
	assertNotContains(t, resp.Output, "👤")
	assertNotContains(t, resp.Output, "──── USER")
	assertNotContains(t, resp.Output, "Role: user")
	// Compact: primary content is single-line style (one prompt → one body line
	// plus optional footer). Fail if output has many blank-separated card blocks.
	if strings.Count(resp.Output, "\n\n\n") > 0 {
		t.Fatalf("unexpected multi-blank card spacing:\n%s", resp.Output)
	}
}
```
