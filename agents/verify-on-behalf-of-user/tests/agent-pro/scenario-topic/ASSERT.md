---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains `name: verify-on-behalf-of-user/scenario`.
- Stdout mentions always labeling depth (`smoke` / `scenario` / `full`).
- Stdout requires browser-agent for UI and forbids playwright-debug for this skill.
- Stdout states missing UI path is FAIL.

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
	if !strings.Contains(resp.Stdout, "name: verify-on-behalf-of-user/scenario") {
		t.Fatalf("missing scenario topic frontmatter:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "browser-agent") {
		t.Fatalf("scenario topic must mention browser-agent:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "playwright-debug") {
		t.Fatalf("scenario topic must call out playwright-debug anti-pattern:\n%s", resp.Stdout)
	}
	if !strings.Contains(strings.ToLower(resp.Stdout), "always label") {
		t.Fatalf("scenario topic must require always labeling depth:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "smoke") || !strings.Contains(resp.Stdout, "scenario") || !strings.Contains(resp.Stdout, "full") {
		t.Fatalf("scenario topic must define smoke|scenario|full depths:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "FAIL") {
		t.Fatalf("scenario topic must state FAIL when UI path broken:\n%s", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unknown skill") || strings.Contains(resp.Stderr, "unknown topic") {
		t.Fatalf("scenario topic not registered:\n%s", resp.Stderr)
	}
}
```
