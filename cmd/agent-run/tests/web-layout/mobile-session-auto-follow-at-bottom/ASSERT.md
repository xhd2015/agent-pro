---
label: e2e, chromium, slow
explanation: Live stream poll ~20s; assistant growth + scroll follow checks
---

## Expected

- Playwright exit code **0**.
- Initial `distanceFromBottom <= 80` on `message-list`.
- While assistant bubble text grows during streaming, `distanceFromBottom` stays `<= 80` on each detected growth step.

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
	if req.Layout != "session-auto-follow-at-bottom" {
		t.Fatalf("expected layout session-auto-follow-at-bottom, got %q", req.Layout)
	}
}
```