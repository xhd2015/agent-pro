---
label: chromium
explanation: playwright negative control for running card
---

## Expected

- `playwright-debug` exits 0.
- `[data-testid="chat-active"]` visible (session page loaded).
- `[data-testid="agent-running-card"]` not present or not visible when status is `idle`.

## Side Effects

- None beyond ephemeral web server lifecycle.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright-debug exit %d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
	if req.Layout != "running-absent" {
		t.Fatalf("expected layout running-absent, got %q", req.Layout)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```