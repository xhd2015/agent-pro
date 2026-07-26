---
label: e2e, ui-automation
explanation: Playwright omits workspace-show-files when listing has no files
---

## Expected

- Playwright exit 0.
- Dir entries visible; `workspace-show-files` not shown.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
