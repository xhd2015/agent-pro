---
label: e2e
---

## Expected

- Exit 0.
- Title `1 shown of 1`.
- `marker-both` present; `marker-alpha-only` and `marker-beta-only` absent.
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

	if !strings.Contains(resp.Stdout, "1 shown of 1") {
		t.Fatalf("title must be 1 shown of 1:\n%s", resp.Stdout)
	}
	assertContains(t, resp.Stdout, "marker-both")
	assertNotContains(t, resp.Stdout, "marker-alpha-only")
	assertNotContains(t, resp.Stdout, "marker-beta-only")
}
```
