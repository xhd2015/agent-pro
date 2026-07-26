---
label: e2e, ui-automation
explanation: Playwright lists and enters dot directories (e.g. .hidden-dir)
---

## Expected

- Playwright exit 0.
- `.hidden-dir` visible as dir; enter updates browse path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
