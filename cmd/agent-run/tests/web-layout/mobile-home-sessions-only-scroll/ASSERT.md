---
label: chromium
explanation: playwright home layout scroll invariants
---

## Expected

- Playwright exit code **0**.
- `document.documentElement.scrollHeight ≈ clientHeight` (±2px); body likewise.
- `[data-testid="session-list"]` has `scrollHeight > clientHeight`.
- After scrolling the session list, `.top-bar-home` and `[data-testid="composer"]` Y positions unchanged (±2px).

## Exit Code

- Playwright process exits 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "home-sessions-only-scroll" {
		t.Fatalf("expected layout home-sessions-only-scroll, got %q", req.Layout)
	}
}
```