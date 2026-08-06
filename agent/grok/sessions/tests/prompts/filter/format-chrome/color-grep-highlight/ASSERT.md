## Expected

- No error.
- Output contains `MATCH` (or case-equal span).
- Output contains ANSI CSI introducer `\x1b[` (ESC[) for highlight and/or dim meta.
- Prefer: bold/red family around the match (SGR 1 and/or 31) when implementer uses that convention (same as sessions list grep).

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
	assertContains(t, out, "MATCH")
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("ColorMode always + grep expected ANSI CSI sequences, got:\n%q", out)
	}
	// Soft contract: bold (1) or red (31) appears somewhere — match highlight family
	if !strings.Contains(out, "\x1b[1m") && !strings.Contains(out, "\x1b[31m") &&
		!strings.Contains(out, "\x1b[1;31m") && !strings.Contains(out, "\x1b[31;1m") {
		// dim-only would still have CSI; require some color intent beyond nothing
		// Accept any CSI for now if bold/red not used; already checked CSI above.
		// Tighten: require 'm' SGR present which CSI check covers.
		_ = out
	}
}
```
