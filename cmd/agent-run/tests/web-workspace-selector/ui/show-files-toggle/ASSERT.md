---
label: e2e, ui-automation
explanation: Playwright expand/collapse workspace-show-files; files disabled when shown
---

## Expected

- Playwright exit 0.
- Expand reveals files (incl. `.env`); files non-selectable; collapse hides again.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
