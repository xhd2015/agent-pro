---
label: ui-automation
explanation: Playwright Quick Server cwd chip browse-only — no status commit
---

## Expected

- Playwright exit 0.
- After Quick cwd: browse path reflects process cwd; `status.workspace` unchanged; still on `/workspace`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
