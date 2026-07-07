---
label: ui-automation
explanation: Playwright poll for progress-card with resolve think text
---

## Expected

- Playwright exits 0 with at least one `[data-testid="progress-card"]` containing
  `Resolve session id`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```