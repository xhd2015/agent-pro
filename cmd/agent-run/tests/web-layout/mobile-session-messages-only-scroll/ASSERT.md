---
label: chromium
explanation: playwright layout scroll invariants
---

## Expected

- Playwright exit code **0**.
- `document.documentElement.scrollHeight ≈ clientHeight` (±2px); body likewise.
- `[data-testid="message-list"]` has `scrollHeight > clientHeight`.
- After scrolling the message list, `.top-bar`, `.session-header`, and `[data-testid="composer"]` Y positions unchanged (±2px).

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
	if req.Layout != "session-messages-only-scroll" {
		t.Fatalf("expected layout session-messages-only-scroll, got %q", req.Layout)
	}
}
```