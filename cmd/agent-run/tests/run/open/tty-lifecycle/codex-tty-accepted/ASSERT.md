---
label: e2e
---

## Expected

- Exit code 0.
- Stderr does **not** treat `codex-tty` as unknown runner.
- Stderr does **not** reject as non-TTY (open is valid for codex-tty).
- Prefer a post-attach `codex-tty: <id>` line when the open path fully lands.

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
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(errText, "unknown") {
		t.Fatalf("codex-tty rejected as unknown runner:\n%s", resp.Stderr)
	}
	if strings.Contains(errText, "non-tty") || strings.Contains(errText, "not a tty") {
		t.Fatalf("codex-tty must accept --open (TTY runner):\n%s", resp.Stderr)
	}
	// Full open path: session id printed after attach.
	if _, ok := parsePrefixedSessionID(resp.Stderr, "codex-tty"); !ok {
		t.Fatalf("expected post-attach codex-tty session id on stderr:\n%s", resp.Stderr)
	}
}
```
