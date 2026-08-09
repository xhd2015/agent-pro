---
label: e2e
---

## Expected

- Exit 0.
- Title includes `2 shown of 4` and `limit 2` (of N = match count, not store size).
- Newest two hits present: `marker-hit-03`, `marker-hit-02`.
- Older hits `marker-hit-00`, `marker-hit-01` absent (limit).
- Miss markers absent.
- `marker-hit-03` before `marker-hit-02`.
- Trailing newline; no ANSI.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertNoANSI(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, "2 shown of 4") || !strings.Contains(resp.Stdout, "limit 2") {
		t.Fatalf("title must be 2 shown of 4, limit 2:\n%s", resp.Stdout)
	}
	// Must not report store size 6 as the "of N" total.
	if strings.Contains(resp.Stdout, "of 6") {
		t.Fatalf("of N must be match count (4), not store size (6):\n%s", resp.Stdout)
	}

	assertContains(t, resp.Stdout, "marker-hit-03")
	assertContains(t, resp.Stdout, "marker-hit-02")
	assertNotContains(t, resp.Stdout, "marker-hit-00")
	assertNotContains(t, resp.Stdout, "marker-hit-01")
	assertNotContains(t, resp.Stdout, "marker-miss-00")
	assertNotContains(t, resp.Stdout, "marker-miss-01")

	i3 := strings.Index(resp.Stdout, "marker-hit-03")
	i2 := strings.Index(resp.Stdout, "marker-hit-02")
	if i3 < 0 || i2 < 0 || i3 > i2 {
		t.Fatalf("expected hit-03 before hit-02; i3=%d i2=%d\n%s", i3, i2, resp.Stdout)
	}
}
```
