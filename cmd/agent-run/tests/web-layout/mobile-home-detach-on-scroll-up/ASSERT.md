---
label: e2e, chromium, slow
explanation: Home poll refresh ~4s; detach + frozen scrollTop checks
---

## Expected

- Playwright exit code **0**.
- After scroll-down from top, `distanceFromTop > 80` (detached).
- After poll refresh adds a 21st session, `session-list.scrollTop` unchanged (±2px).

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
	if req.Layout != "home-detach-on-scroll-up" {
		t.Fatalf("expected layout home-detach-on-scroll-up, got %q", req.Layout)
	}
}
```
