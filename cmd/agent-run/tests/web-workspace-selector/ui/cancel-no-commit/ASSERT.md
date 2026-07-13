---
label: ui-automation
explanation: Playwright Cancel returns home without PUT
---

## Expected

- Playwright exit 0.
- After Cancel: home URL; `status.workspace` still original selected path.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	cfg := readHomeConfigMap(t, req.Home)
	if cfg != nil {
		if sel, _ := cfg["selected_workspace"].(string); sel != "" && !pathsEqual(sel, req.SelectPath) {
			t.Fatalf("config selected_workspace changed after Cancel: got %q want %q", sel, req.SelectPath)
		}
	}
}
```
