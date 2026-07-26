---
label: e2e, ui-automation
explanation: Playwright re-hides files when browse path changes after expand
---

## Expected

- Playwright exit 0.
- After expand + enter `src`, files hidden and `workspace-show-files` collapsed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
