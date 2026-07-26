---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `name: verify-on-behalf-of-user/tty`.
- Stdout requires `tty-watch` and `run --detach` (non-blocking).
- Stdout requires `tty-watch kill` to reclaim resources.
- Stdout mentions `snapshot` and/or `send`.
- Stdout forbids relying on pipe-only / raw openpty as sole interactive proof.

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
	assertExitCode(t, resp, 0)
	if !strings.Contains(resp.Stdout, "name: verify-on-behalf-of-user/tty") {
		t.Fatalf("missing tty topic frontmatter:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "tty-watch") {
		t.Fatalf("tty topic must mention tty-watch:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "--detach") {
		t.Fatalf("tty topic must require run --detach:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "kill") {
		t.Fatalf("tty topic must require tty-watch kill to reclaim:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "snapshot") && !strings.Contains(resp.Stdout, "send") {
		t.Fatalf("tty topic must mention snapshot and/or send:\n%s", resp.Stdout)
	}
	lower := strings.ToLower(resp.Stdout)
	if !strings.Contains(lower, "openpty") && !strings.Contains(lower, "pipe") {
		t.Fatalf("tty topic should call out openpty/pipe anti-patterns:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unknown skill") || strings.Contains(resp.Stderr, "unknown topic") {
		t.Fatalf("tty topic not registered:\n%s", resp.Stderr)
	}
}
```
