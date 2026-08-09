---
label: e2e
---

## Expected

- Exit 0.
- Title `1 shown of 1`.
- `marker-cross` present; `marker-orphan` absent.
- Both Q alpha and A beta lines of the cross session appear.
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
	assertContains(t, resp.Stdout, "marker-cross")
	assertContains(t, resp.Stdout, "talk about alpha marker-cross")
	assertContains(t, resp.Stdout, "and about beta in the answer")
	assertNotContains(t, resp.Stdout, "marker-orphan")
}
```
