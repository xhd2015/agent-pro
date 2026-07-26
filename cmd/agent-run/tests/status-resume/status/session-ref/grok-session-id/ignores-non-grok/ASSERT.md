---
label: e2e
---

## Expected

- Exit code 1 (not found — codex must not resolve).
- Combined output indicates not found / no match.
- Must **not** successfully print multi-layer status for `test-gsid-codex-s1`
  (exit 0 would be a fail).

## Exit Code

1

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
	assertExitCode(t, resp, 1)
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"not found",
		"no such",
		"unknown",
		"no match",
		"no session",
	)
	// If a bad implementation resolved codex, it would exit 0 — already checked.
	// Guard against listing the codex agent-run id as a successful status line.
	if resp.ExitCode == 0 && strings.Contains(resp.Stdout, req.SessionID) {
		t.Fatalf("must not resolve non-grok session %q via --grok-session-id", req.SessionID)
	}
}
```
