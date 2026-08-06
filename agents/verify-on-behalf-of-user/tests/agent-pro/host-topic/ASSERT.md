---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `name: verify-on-behalf-of-user/host`.
- Stdout states default is sandbox / host is opt-in only.
- Stdout mentions `wrk --reinstall-local` and `--dry-run`.
- Stdout mentions change-scoped targets / script install / go install fallbacks.
- Stdout requires `warning:` for host mode.

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
	if !strings.Contains(resp.Stdout, "name: verify-on-behalf-of-user/host") {
		t.Fatalf("missing host topic frontmatter:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	lower := strings.ToLower(resp.Stdout)
	if !strings.Contains(lower, "sandbox") || !strings.Contains(lower, "default") {
		t.Fatalf("host topic must state sandbox is default:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "outside sandbox") && !strings.Contains(lower, "opt-in") {
		t.Fatalf("host topic must describe opt-in / outside sandbox:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "wrk --reinstall-local") {
		t.Fatalf("host topic must mention wrk --reinstall-local:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "--dry-run") {
		t.Fatalf("host topic must mention --dry-run:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "go install") {
		t.Fatalf("host topic must mention go install fallback:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "script/") && !strings.Contains(resp.Stdout, "./script") {
		t.Fatalf("host topic must mention script install patterns:\n%s", resp.Stdout)
	}
	// change-scoped: cmd/T or my-tool style examples
	if !strings.Contains(resp.Stdout, "./cmd/") && !strings.Contains(resp.Stdout, "cmd/T") && !strings.Contains(resp.Stdout, "my-tool") {
		t.Fatalf("host topic must describe change-scoped cmd/script targets:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "warning:") {
		t.Fatalf("host topic must require warning: lines:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unknown skill") || strings.Contains(resp.Stderr, "unknown topic") {
		t.Fatalf("host topic not registered:\n%s", resp.Stderr)
	}
}
```
