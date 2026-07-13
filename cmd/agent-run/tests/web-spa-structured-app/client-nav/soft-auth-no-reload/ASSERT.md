---
label: ui-automation
explanation: Playwright soft auth without full document reload
---

## Expected

- Playwright exits 0.
- After submit: home/empty UI visible; auth-page gone.
- `window.__SPA_NAV_MARKER === 'alive'` — soft auth (no `window.location` hard reload).

## Side Effects

- Background web cleanup.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
