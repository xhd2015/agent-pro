## Expected

- Exit code 0.
- Stdout documents session status usage (session id and/or `runner/session` ref).
- Mentions multi-layer concepts when documented (session / process / terminal /
  runner / resume) — at least session id or session-ref wording is required.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	assertTrailingNewline(t, resp.Stdout, "status help stdout")
}
```
