---
label: ui-automation
explanation: Playwright re-hides files when browse path changes after expand
---

## Expected

- Playwright exit 0.
- After expand + enter `src`, files hidden and `workspace-show-files` collapsed.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
