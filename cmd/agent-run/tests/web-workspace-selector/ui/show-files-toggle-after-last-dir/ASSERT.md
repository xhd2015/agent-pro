---
label: e2e, ui-automation
explanation: Playwright document-order of dirs → workspace-show-files → files
---

## Expected

- Playwright exit 0.
- Collapsed document order under browse section: one or more `dir`, then
  `workspace-show-files`, no `file`.
- Expanded: all `dir`, then `workspace-show-files`, then one or more `file`
  (first file after the toggle in document order).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
