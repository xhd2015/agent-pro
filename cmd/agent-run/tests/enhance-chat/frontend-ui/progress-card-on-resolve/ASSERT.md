---
label: e2e, ui-automation
explanation: Playwright poll for progress-card with resolve think text
---

## Expected

- Playwright exits 0 with at least one `[data-testid="progress-card"]` containing
  `Resolve session id`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```