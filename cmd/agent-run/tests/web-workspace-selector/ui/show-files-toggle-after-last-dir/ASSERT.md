---
label: ui-automation
explanation: Playwright document-order of dirs → workspace-show-files → files
---

## Expected

- Playwright exit 0.
- Collapsed document order under browse section: one or more `dir`, then
  `workspace-show-files`, no `file`.
- Expanded: all `dir`, then `workspace-show-files`, then one or more `file`
  (first file after the toggle in document order).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
