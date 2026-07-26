---
label: e2e, ui-automation
explanation: Requires playwright-debug and browser automation.
---

## Expected

- Playwright exits 0.
- Closing the terminal modal after the first tty response does not leave an old
  SSE stream that can replay the next user message.
- The follow-up user message is visible exactly once after the post-send
  refresh and stream startup windows.

## Exit Code

- Test process exits non-zero until stale SSE streams are aborted or ignored
  when the refreshed event offset changes.

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
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s", resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
