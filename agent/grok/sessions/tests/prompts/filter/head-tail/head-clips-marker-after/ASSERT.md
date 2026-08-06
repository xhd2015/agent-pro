## Expected

- No error.
- Structured: 2 prompts p1,p2; OmittedAfter=3; OmittedBefore=0.
- Output contains p1 and p2; does not contain p3/p4/p5 as prompt lines.
- Output contains exact `(...3 omitted...)`.
- Marker appears **after** the printed prompt lines (last content line or after p2).
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
	sp := resp.Single
	assertPromptCount(t, sp, 2)
	assertPromptText(t, sp, 0, "p1")
	assertPromptText(t, sp, 1, "p2")
	if sp.OmittedAfter != 3 {
		t.Fatalf("OmittedAfter=%d want 3", sp.OmittedAfter)
	}
	if sp.OmittedBefore != 0 {
		t.Fatalf("OmittedBefore=%d want 0", sp.OmittedBefore)
	}
	out := resp.Output
	assertTrailingNewline(t, out)
	assertContains(t, out, "p1")
	assertContains(t, out, "p2")
	assertContains(t, out, omissionMarker(3))
	// p3-p5 should not appear as full prompt bodies (marker is not a prompt)
	for _, drop := range []string{"p3", "p4", "p5"} {
		// allow if somehow inside marker text — marker only has digits
		if strings.Contains(out, drop) {
			t.Fatalf("output still contains dropped prompt %q:\n%s", drop, out)
		}
	}
	// marker after prompt lines: last non-empty line is the marker
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("want >=3 lines, got %d:\n%s", len(lines), out)
	}
	if lines[len(lines)-1] != omissionMarker(3) {
		t.Fatalf("last line want %q, got %q\nfull:\n%s", omissionMarker(3), lines[len(lines)-1], out)
	}
}
```
