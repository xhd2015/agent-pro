---
label: chromium
explanation: playwright
---

## Expected

- `playwright-debug` exits 0.
- Viewport 390×844; no horizontal document scroll.
- `[data-testid="auth-page"]` visible (user prompted for token after 401).
- Token input (`[data-testid="auth-token-input"]` or first input in auth page) anchored near viewport bottom (≤80px gap).

## Side Effects

- Background `agent-run web` process started during Setup and stopped on test cleanup.

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
	if req.Layout != "auth" {
		t.Fatalf("expected layout auth, got %q", req.Layout)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```