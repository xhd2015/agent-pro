---
label: e2e, chromium
explanation: playwright; workspace header and role-specific message testids
---

## Expected

- `playwright-debug` exits 0.
- `[data-testid="workspace"]` visible and contains seeded workspace path.
- `[data-testid="message-item-user"]` visible with user prompt text.
- `[data-testid="message-item-assistant"]` visible.
- Composer remains pinned to viewport bottom.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright-debug exit %d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
	if req.Layout != "workspace-session" {
		t.Fatalf("expected layout workspace-session, got %q", req.Layout)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```