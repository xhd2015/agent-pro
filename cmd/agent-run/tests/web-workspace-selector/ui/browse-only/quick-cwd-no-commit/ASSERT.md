---
label: e2e, ui-automation
explanation: Playwright Quick Server cwd chip browse-only — no status commit
---

## Expected

- Playwright exit 0.
- After Quick cwd: browse path reflects process cwd; `status.workspace` unchanged; still on `/workspace`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
