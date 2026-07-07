---
label: ui-automation
explanation: Playwright poll for error-card on bind failure; no assistant fallback bubble
---

## Expected

- Playwright exits 0 with visible `[data-testid="error-card"]` containing
  `Cannot resolve session id`.
- No `[data-testid="assistant-message"]` bubble from PTY scrollback fallback.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```