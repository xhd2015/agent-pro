---
label: ui-automation
explanation: Playwright session back-link to home
---

## Expected

- Playwright exits 0.
- After back: URL `/` and home UI visible (`home-active` / `empty-state` / `session-list`).
- Nav marker survives (soft client navigation).

## Side Effects

- Seeded session; web cleanup.

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
