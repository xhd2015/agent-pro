---
label: ui-automation
explanation: Playwright expand/collapse workspace-show-files; files disabled when shown
---

## Expected

- Playwright exit 0.
- Expand reveals files (incl. `.env`); files non-selectable; collapse hides again.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
