---
label: e2e, ui-automation
explanation: Playwright Recent row browse-only — no status commit
---

## Expected

- Playwright exit 0.
- Recent tap updates browse path only; `status.workspace` unchanged; still on `/workspace`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
