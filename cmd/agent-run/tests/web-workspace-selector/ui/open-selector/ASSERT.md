---
label: ui-automation
explanation: Playwright home opens /workspace selector page
---

## Expected

- `playwright-debug` exits 0.
- After click: URL is `/workspace`.
- `[data-testid="workspace-selector"]` and `workspace-use-folder` visible.

## Errors

- Pre-impl: workspace control does not navigate / selector missing (RED).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "open-selector" {
		t.Fatalf("expected scenario open-selector, got %q", req.Scenario)
	}
}
```
