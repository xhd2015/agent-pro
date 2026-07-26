---
label: e2e, ui-automation
explanation: Playwright asserts files hidden by default; show-files control present collapsed
---

## Expected

- Playwright exit 0.
- Dir entries visible; no file entries; `workspace-show-files` visible and not expanded.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
