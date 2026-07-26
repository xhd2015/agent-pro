---
label: e2e, ui-automation
explanation: Playwright opens terminal modal on web-created grok-tty session
---

## Expected

- Terminal modal renders xterm surface with CODEX_TTY_BANNER scrollback from attach relay.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
