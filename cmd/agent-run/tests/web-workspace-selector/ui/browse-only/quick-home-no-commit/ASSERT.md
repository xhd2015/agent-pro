---
label: e2e, ui-automation
explanation: Playwright Quick Home chip browse-only — no status commit
---

## Expected

- Playwright exit 0.
- After Quick Home: browse path updated; `status.workspace` still selected; still on `/workspace`.

## Errors

- Pre-impl: selector/chip missing or chip auto-commits (RED).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
