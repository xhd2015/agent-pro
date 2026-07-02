---
label: ui-automation
explanation: Requires playwright-debug and browser automation.
---

## Expected

- Playwright exits 0.
- Clicking Terminal during an active `codex-tty` turn attaches to the live PTY.
- The modal shows initial terminal output from the running process.
- The modal does not show `terminal unavailable`.
- The modal does not show an exited terminal before the active turn finishes.

## Exit Code

- Test process exits non-zero until the chat page can attach to a live running
  tty session before assistant response completion.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d\nstdout:\n%s\nstderr:\n%s", resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
}
```
