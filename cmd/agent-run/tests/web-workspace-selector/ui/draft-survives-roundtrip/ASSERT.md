---
label: e2e, ui-automation
explanation: Playwright composer draft survives /workspace Cancel round-trip
---

## Expected

- Playwright exit 0.
- After Cancel: composer input still contains typed draft.

## Errors

- Pre-impl: no /workspace route and/or local draft state wiped (RED).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
