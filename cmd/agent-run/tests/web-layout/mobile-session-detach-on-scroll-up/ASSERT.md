---
label: e2e, chromium, slow
explanation: Live stream poll ~20s; detach + frozen scrollTop checks
---

## Expected

- Playwright exit code **0**.
- After scroll-up, `distanceFromBottom > 80` (detached).
- While assistant text grows, `message-list.scrollTop` unchanged (±2px).

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
	if req.Layout != "session-detach-on-scroll-up" {
		t.Fatalf("expected layout session-detach-on-scroll-up, got %q", req.Layout)
	}
}
```