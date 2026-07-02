---
label: ui-automation
explanation: Uses playwright-debug against a generated web-created codex-tty session.
---

## Expected

- Playwright exits 0.
- Browser-side terminal API call reports `available:true`.
- Browser-side terminal API call echoes the generated `terminal_session_id`.
- Generated finished codex-tty chat page shows a visible enabled terminal button.

## Side Effects

- Browser localStorage stores the test auth token.
- Creates one isolated codex-tty web session in the test home.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero while newly created codex-tty web sessions do not
  keep their backend terminal attachable after the first turn.

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
