---
label: e2e, chromium, slow
explanation: Playwright network monitor; delayed registry + 8s discovery window
---

## Expected

- Playwright exit code **0**.
- Running `grok-tty` session without `terminal_session_id` in session detail:
  - Registry appears after **~1s** delay.
  - `GET .../terminal` count is **> 0** during discovery.
  - Total terminal GET count in **8s** is **≤ 8** (not 16+ from 500ms perpetual poll).
  - **Zero** terminal GETs after first `available: true` response (poll stops).
  - No ~500ms interval between consecutive terminal GETs.

## Side Effects

- Delayed registry file written under `AGENT_RUN_HOME/grok-tty-registry/`.

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
	if req.Scenario != "discovery-poll-stops-after-available" {
		t.Fatalf("expected scenario discovery-poll-stops-after-available, got %q", req.Scenario)
	}
}
```