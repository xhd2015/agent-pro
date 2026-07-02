---
label: ui-automation
explanation: Requires playwright-debug and browser automation.
---

## Expected

- Playwright exits 0.
- After the terminal is closed, the chat follow-up is delivered to the existing
  backend TTY and the resulting assistant output is surfaced back to the chat.
- The chat page shows `FOLLOWUP_RESPONSE`, proving the fake two-turn Codex
  process read the follow-up as a submitted second line and the web session
  streamed/persisted the response.

## Exit Code

- Test process exits non-zero while the follow-up does not produce a visible
  chat response after the terminal modal has been closed.

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
