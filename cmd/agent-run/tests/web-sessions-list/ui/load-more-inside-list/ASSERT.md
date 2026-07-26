---
label: e2e, ui-automation
explanation: Playwright asserts session-load-more is inside session-list
---

## Expected

- `playwright-debug` exits **0**.
- `[data-testid="session-load-more"]` is contained by `[data-testid="session-list"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "load-more-inside-list" {
		t.Fatalf("expected scenario load-more-inside-list, got %q", req.Scenario)
	}
}
```
