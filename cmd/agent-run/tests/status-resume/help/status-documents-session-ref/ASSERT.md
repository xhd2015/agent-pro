---
label: e2e
---

## Expected

- Exit code 0.
- Stdout documents session status usage (session id and/or `runner/session` ref).
- Mentions multi-layer concepts when documented (session / process / terminal /
  runner / resume) — at least session id or session-ref wording is required.
- Stdout documents `--grok-session-id`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

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
	assertSuccess(t, resp)
	out := strings.ToLower(resp.Stdout)
	assertContains(t, out, "status")
	// Session-ref documentation: any of these phrases satisfy H2.
	assertContainsAny(t, out,
		"session",
		"session-id",
		"session id",
		"<session",
		"runner/",
	)
	assertContains(t, resp.Stdout, "--grok-session-id")
	assertTrailingNewline(t, resp.Stdout, "status help stdout")
}
```
