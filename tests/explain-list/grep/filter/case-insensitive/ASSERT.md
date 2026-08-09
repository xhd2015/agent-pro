---
label: e2e
---

## Expected

- Exit 0.
- Title `1 shown of 1`.
- Body contains original `Docker` (capital D).
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
		t.Fatalf("expected match listed:\n%s", resp.Stdout)
	}
	// Original casing preserved (pattern was lowercase "docker").
	assertContains(t, resp.Stdout, "Docker")
	assertContains(t, resp.Stdout, "How does Docker networking work?")
}
```
