---
label: ui-automation
explanation: Playwright asserts scroll near bottom does not auto-paginate
---

## Expected

- `playwright-debug` exits **0**.
- After scroll-to-bottom without click, item count stays at first-page size for ≥800ms.
- No `offset>0` sessions GET triggered by scroll alone.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "scroll-does-not-auto-load" {
		t.Fatalf("expected scenario scroll-does-not-auto-load, got %q", req.Scenario)
	}
}
```
