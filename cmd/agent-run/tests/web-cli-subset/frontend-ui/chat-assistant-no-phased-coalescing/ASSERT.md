---
label: e2e, ui-automation
explanation: Playwright chat DOM poll for assistant bubble count
---

## Expected

- Playwright exits 0 with exactly one assistant bubble and no phased DOM rows.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
