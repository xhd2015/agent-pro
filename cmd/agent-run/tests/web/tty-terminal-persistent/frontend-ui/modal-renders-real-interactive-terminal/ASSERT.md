---
label: ui-automation
explanation: Requires playwright-debug and browser automation.
---

## Expected

- Playwright exits 0.
- Terminal modal contains a real terminal emulator surface.
- ANSI control bytes are interpreted and not shown as raw text.
- `session_id` control JSON is hidden from the user.
- Typed input reaches the backend PTY websocket and echoed output is rendered.

## Exit Code

- Test process exits non-zero until the modal renders an interactive terminal
  emulator instead of a raw websocket transcript.

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
