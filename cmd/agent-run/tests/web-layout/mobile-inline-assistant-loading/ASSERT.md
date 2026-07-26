---
label: e2e, chromium
explanation: Playwright mobile viewport; seeded running session DOM
---

## Expected

- Playwright exit code **0**.
- Script verified `message-item-assistant-loading` is visible and positioned below the last user bubble.

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
}
```