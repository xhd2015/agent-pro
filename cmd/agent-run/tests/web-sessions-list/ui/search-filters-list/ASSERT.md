---
label: ui-automation
explanation: Playwright home session-search filters list
---

## Expected

- `playwright-debug` exits **0**.
- Typing into `session-search` leaves a single matching row; clear restores more.

## Errors

- Pre-impl: missing `session-search` control (RED).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "search-filters-list" {
		t.Fatalf("expected scenario search-filters-list, got %q", req.Scenario)
	}
}
```
