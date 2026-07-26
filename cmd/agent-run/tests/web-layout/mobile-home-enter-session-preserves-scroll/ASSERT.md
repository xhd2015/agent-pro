---
label: e2e, ui-automation
explanation: enter session then back preserves home list scrollTop
---

## Expected

- Playwright exit code **0**.
- After session → back, `session-list.scrollTop` within ~60px of pre-nav.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "home-enter-session-preserves-scroll" {
		t.Fatalf("expected layout home-enter-session-preserves-scroll, got %q", req.Layout)
	}
}
```
