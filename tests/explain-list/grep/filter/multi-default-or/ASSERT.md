---
label: e2e
---

## Expected

- Exit 0.
- Title `2 shown of 2`.
- `marker-alpha` and `marker-beta` present; `marker-gamma` absent.
- `marker-beta` appears before `marker-alpha` (newer first).
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

	if !strings.Contains(resp.Stdout, "2 shown of 2") {
		t.Fatalf("title must be 2 shown of 2:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "marker-alpha")
	assertContains(t, resp.Stdout, "marker-beta")
	assertNotContains(t, resp.Stdout, "marker-gamma")

	ib := strings.Index(resp.Stdout, "marker-beta")
	ia := strings.Index(resp.Stdout, "marker-alpha")
	if ib < 0 || ia < 0 || ib > ia {
		t.Fatalf("expected beta (newer) before alpha; ib=%d ia=%d\n%s", ib, ia, resp.Stdout)
	}
}
```
