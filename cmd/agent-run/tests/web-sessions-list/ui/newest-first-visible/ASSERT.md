---
label: e2e, ui-automation
explanation: Playwright home session list newest-first order
---

## Expected

- `playwright-debug` exits **0**.
- First `session-item` preview matches newest seed (`brand newest epsilon`).

## Errors

- Pre-impl: oldest-first sort puts alpha first (RED).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "newest-first-visible" {
		t.Fatalf("expected scenario newest-first-visible, got %q", req.Scenario)
	}
}
```
