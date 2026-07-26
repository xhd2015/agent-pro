---
label: e2e, ui-automation
explanation: Uses playwright-debug against a generated web-created grok-tty session.
---

## Expected

- Playwright exits 0.
- Browser-side terminal API call reports `available:true`.
- Browser-side terminal API call echoes the generated `terminal_session_id`.
- Generated finished grok-tty chat page shows a visible enabled terminal button.

## Side Effects

- Browser localStorage stores the test auth token.
- Creates one isolated grok-tty web session in the test home.

## Errors

- None from `Run`.

## Exit Code

- Test process exits non-zero while newly created codex-tty web sessions do not
  keep their backend terminal attachable after the first turn.

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
