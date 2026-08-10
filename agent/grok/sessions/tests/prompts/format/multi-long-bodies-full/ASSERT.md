## Expected

- No error.
- Output contains both full long bodies (`A-` + 220 x's and `B-` + 220 x's).
- No body-cap `…` on those lines.
- Session headers and footer still present (shape unchanged).
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
	out := resp.Output
	assertTrailingNewline(t, out)

	fullA := "A-" + longPromptRunes(220)
	fullB := "B-" + longPromptRunes(220)
	if !strings.Contains(out, fullA) {
		t.Fatalf("multi format missing full body for session A:\n%s", out)
	}
	if !strings.Contains(out, fullB) {
		t.Fatalf("multi format missing full body for session B:\n%s", out)
	}
	// Soft-cap would introduce … on truncated lines; default must not.
	if strings.Contains(out, "…") {
		t.Fatalf("default multi long bodies must not soft-cap with …:\n%s", out)
	}
	assertContains(t, out, idFormatMultiA)
	assertContains(t, out, idFormatMultiB)
}
```
