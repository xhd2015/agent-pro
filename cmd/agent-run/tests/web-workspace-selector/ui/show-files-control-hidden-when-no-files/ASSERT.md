---
label: ui-automation
explanation: Playwright omits workspace-show-files when listing has no files
---

## Expected

- Playwright exit 0.
- Dir entries visible; `workspace-show-files` not shown.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
