---
label: e2e
---

## Expected

- Exit code 0 (help is success, even if resume is newly added).
- Stdout contains `--open`.
- Stdout mentions session id / followup positional usage.
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
	assertContains(t, resp.Stdout, "--open")
	out := strings.ToLower(resp.Stdout)
	assertContainsAny(t, out,
		"session",
		"session-id",
		"session id",
		"<session",
	)
	assertContains(t, resp.Stdout, "--grok-session-id")
	assertTrailingNewline(t, resp.Stdout, "resume help stdout")
}
```
