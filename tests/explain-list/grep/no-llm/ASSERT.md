---
label: e2e
---

## Expected

- Exit 0.
- Stdout lists the seeded session (`marker-nollm`).
- Stderr does not contain `FAKE_AGENT_INVOKED`.
- Exit code is not 99 (fake agent exit).
- Trailing newline; no ANSI.

## Side Effects

- None (read-only).

## Errors

- None.

## Exit Code

- 0.

```go
import (
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
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")
	assertNotContains(t, resp.Stdout, "FAKE_AGENT_INVOKED")
	assertContains(t, resp.Stdout, "marker-nollm")
}
```
