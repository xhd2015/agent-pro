---
label: e2e
---

## Expected

- Exit 0.
- Title includes `3 shown of 3` and `limit 100` (capped).
- Does not report `limit 200`.
- All three questions present.
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

	if !strings.Contains(resp.Stdout, "3 shown of 3") || !strings.Contains(resp.Stdout, "limit 100") {
		t.Fatalf("expected 3 shown and limit 100 (cap):\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stdout, "limit 200") {
		t.Fatalf("limit must be capped at 100, not 200:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "question-00")
	assertContains(t, resp.Stdout, "question-01")
	assertContains(t, resp.Stdout, "question-02")
}
```
