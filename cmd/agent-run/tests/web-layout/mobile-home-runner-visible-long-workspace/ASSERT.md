---
label: chromium
explanation: playwright mobile home header runner visibility
---

## Expected

- `playwright-debug` exits 0.
- Viewport 390×844; `documentElement.scrollWidth <= clientWidth` (no horizontal scroll).
- `[data-testid="workspace"]` shows a long server cwd string (may truncate visually).
- `[data-testid="runner-picker"]` bounding box lies within viewport width (right edge ≤ 390px).
- `[data-testid="runner-select"]` visible.
- `[data-testid="empty-state"]` and `[data-testid="composer"]` still present.

## Side Effects

- Web server runs with `cmd.Dir` set to the deep workspace path for the duration of the test.

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
	if req.Layout != "home-long-workspace" {
		t.Fatalf("expected layout home-long-workspace, got %q", req.Layout)
	}
	if req.WebWorkingDir == "" {
		t.Fatal("WebWorkingDir must be set for long workspace home test")
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```