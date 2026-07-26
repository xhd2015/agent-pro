---
label: e2e, ui-automation
explanation: Playwright home page-size 30 + explicit load-more button appends
---

## Expected

- `playwright-debug` exits **0**.
- Initial visible rows ≤ **30**; after **button** click, count increases.
- Does not rely on scroll-near-bottom auto-load.

## Errors

- Missing `session-load-more` or click does not append → fail.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "load-more-appends" {
		t.Fatalf("expected scenario load-more-appends, got %q", req.Scenario)
	}
}
```
