---
label: ui-automation
explanation: Playwright client-nav home→session; no full reload marker
---

## Expected

- `playwright-debug` exits 0.
- After click: URL is `/sessions/spa-nav-home-sess` and `[data-testid="chat-active"]` visible.
- `window.__SPA_NAV_MARKER === 'alive'` (no full document reload).

## Side Effects

- Seeded session dir; background web cleanup.

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
