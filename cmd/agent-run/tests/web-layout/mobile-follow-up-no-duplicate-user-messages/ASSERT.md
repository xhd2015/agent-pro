---
label: e2e, chromium
explanation: Playwright mobile follow-up; live grok-tty run ~60s
---

## Expected

- Playwright exit code **0**.
- Viewport 390×844; no horizontal document scroll.
- **During** the follow-up run (while `agent-running-card`, inline loading, or `running` status pill is visible), `[data-testid="message-item-user"]` count never exceeds **2** at any 250ms poll step.
- After follow-up, poll until `[data-testid="message-item-user"]` count is **exactly 2** (does **not** require session idle/finished or assistant bubbles).
- Initial prompt `first layout prompt` and follow-up `second follow-up prompt` each appear in exactly one user bubble.

## Side Effects

- Session files under flat `AGENT_RUN_HOME/sessions/follow-up-dedupe/`.
- Follow-up POST enqueues agent run via `grok-tty` mock harness.

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
	if req.Layout != "follow-up-dedupe" {
		t.Fatalf("expected layout follow-up-dedupe, got %q", req.Layout)
	}
}
```
