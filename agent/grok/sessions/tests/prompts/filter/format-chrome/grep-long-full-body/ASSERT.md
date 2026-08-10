## Expected

- No error.
- Full collapsed body present: 150 leading `a`, 150 trailing `b`, and `MATCH`.
- No body-window / soft-cap `…` (U+2026) when MaxBody unset.
- ColorMode always: ANSI CSI around the match (bold/red family preferred).
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

	as := strings.Repeat("a", 150)
	bs := strings.Repeat("b", 150)
	if !strings.Contains(out, as) {
		t.Fatalf("grep+full default missing leading 150 a's (windowed/truncated?):\n%s", out)
	}
	if !strings.Contains(out, bs) {
		t.Fatalf("grep+full default missing trailing 150 b's:\n%s", out)
	}
	if !strings.Contains(out, "MATCH") {
		t.Fatalf("expected MATCH in output:\n%s", out)
	}
	if strings.Contains(out, "…") {
		t.Fatalf("no MaxBody: must not window/soft-cap with …:\n%s", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("ColorMode always + grep expected ANSI CSI, got:\n%q", out)
	}
}
```
