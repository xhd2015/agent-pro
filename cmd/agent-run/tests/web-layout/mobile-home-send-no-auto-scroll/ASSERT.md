---
label: e2e, chromium
explanation: playwright home detach + composer send scroll position
---

## Expected

- Playwright exit code **0**.
- After scroll-up detach, `distanceFromBottom > 80`.
- Within 500ms after composer send, `session-list.scrollTop` unchanged (±2px).

## Exit Code

- Playwright process exits 0.

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
	if req.Layout != "home-send-no-auto-scroll" {
		t.Fatalf("expected layout home-send-no-auto-scroll, got %q", req.Layout)
	}
}
```