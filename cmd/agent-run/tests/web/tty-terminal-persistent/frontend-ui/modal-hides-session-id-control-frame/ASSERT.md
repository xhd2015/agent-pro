---
label: ui-automation
explanation: Uses playwright-debug to verify modal websocket rendering.
---

## Expected

- Playwright exits 0.
- Terminal modal displays the mapped PTY transcript `mapped-terminal-ready`.
- Terminal modal does not display ptywrap `session_id` control JSON.
- Terminal websocket attaches to the mapped `terminal_session_id`, not an
  implicit newly-created ptywrap shell.

## Side Effects

- Browser localStorage stores the test auth token.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero while the modal renders ptywrap control JSON or
  the proxy attaches without the mapped terminal id.

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
