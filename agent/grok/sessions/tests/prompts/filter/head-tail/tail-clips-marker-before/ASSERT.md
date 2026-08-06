## Expected

- No error.
- Structured: 2 prompts p4,p5; OmittedBefore=3; OmittedAfter=0.
- Output contains exact `(...3 omitted...)` **before** p4/p5 lines.
- p1,p2,p3 absent.

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
	assertPromptText(t, sp, 0, "p4")
	assertPromptText(t, sp, 1, "p5")
	if sp.OmittedBefore != 3 {
		t.Fatalf("OmittedBefore=%d want 3", sp.OmittedBefore)
	}
	if sp.OmittedAfter != 0 {
		t.Fatalf("OmittedAfter=%d want 0", sp.OmittedAfter)
	}
	out := resp.Output
	assertTrailingNewline(t, out)
	assertContains(t, out, omissionMarker(3))
	assertContains(t, out, "p4")
	assertContains(t, out, "p5")
	for _, drop := range []string{"p1", "p2", "p3"} {
		if strings.Contains(out, drop) {
			t.Fatalf("output still contains dropped prompt %q:\n%s", drop, out)
		}
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	// first non-empty line should be marker
	if len(lines) < 3 {
		t.Fatalf("want >=3 lines, got %d:\n%s", len(lines), out)
	}
	if lines[0] != omissionMarker(3) {
		t.Fatalf("first line want marker %q, got %q\nfull:\n%s", omissionMarker(3), lines[0], out)
	}
}
```
