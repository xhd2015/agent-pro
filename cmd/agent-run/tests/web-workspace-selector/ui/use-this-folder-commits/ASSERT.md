---
label: e2e, ui-automation
explanation: Playwright Use this folder commits PUT and returns home
---

## Expected

- Playwright exit 0.
- After Use: URL `/`; status.workspace = target; home label reflects selection.

## Errors

- Pre-impl: Use CTA / PUT missing (RED).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	// Optional disk check if server wrote config.
	cfg := readHomeConfigMap(t, req.Home)
	if cfg != nil {
		if sel, _ := cfg["selected_workspace"].(string); sel != "" && !pathsEqual(sel, req.SelectPath) {
			t.Fatalf("config selected_workspace after Use: got %q want %q", sel, req.SelectPath)
		}
	}
}
```
