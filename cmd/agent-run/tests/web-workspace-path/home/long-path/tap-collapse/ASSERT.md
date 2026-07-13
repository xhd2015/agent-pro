---
label: ui-automation
explanation: playwright mobile home workspace expand then collapse (expect RED pre-impl)
---

## Expected

- Expand then second tap on `[data-testid="workspace-toggle"]` returns short `…/` label.
- `aria-expanded="false"` after collapse.
- Runner-picker remains within 390px viewport.
- No horizontal document scroll.

## Errors

- Pre-impl: missing toggle → RED.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "home-long-tap-collapse" {
		t.Fatalf("expected scenario home-long-tap-collapse, got %q", req.Scenario)
	}
}
```
