---
label: e2e, chromium, slow
explanation: Playwright network monitor; 8s passive watch for terminal GET bounds
---

## Expected

- Playwright exit code **0**.
- Viewport 390×844.
- During **8s** passive watch on a finished `grok-tty` session whose detail already
  includes `terminal_session_id` and live registry:
  - `GET .../terminal` count is **≤ 1** (prefer 0 when detail suffices; at most one
    optional mount probe).
  - No ~500ms or ~5s repeating terminal GET pattern.

## Side Effects

- Background `agent-run web` serves seeded session and terminal status endpoints.

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
	if req.Scenario != "known-terminal-id-no-repeat-poll" {
		t.Fatalf("expected scenario known-terminal-id-no-repeat-poll, got %q", req.Scenario)
	}
}
```