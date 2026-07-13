---
label: ui-automation
explanation: Playwright asserts files hidden by default; show-files control present collapsed
---

## Expected

- Playwright exit 0.
- Dir entries visible; no file entries; `workspace-show-files` visible and not expanded.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
