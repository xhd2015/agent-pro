---
label: e2e, chromium
explanation: playwright
---

## Expected

- `playwright-debug` exits 0.
- Viewport 390×844; no horizontal document scroll.
- `[data-testid="chat-active"]` and `[data-testid="message-list"]` visible.
- At least one `[data-testid="message-item"]` rendered.
- `[data-testid="composer"]` visible and pinned to the bottom of the viewport (≤4px gap).

## Side Effects

- Session files written under `AGENT_RUN_HOME/sessions/fake-opencode/layout-chat/`.
- Background `agent-run web` process started during Setup and stopped on test cleanup.

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
	if req.Layout != "chat-active" {
		t.Fatalf("expected layout chat-active, got %q", req.Layout)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```