---
label: chromium, slow
explanation: Live grok-tty run; 15s passive watch — session-detail GET must stay at initial snapshot only
---

## Expected

- Playwright exit code **0**.
- Viewport 390×844; no horizontal document scroll.
- During the **15s** passive monitoring window after opening a live `grok-tty` session (no composer interaction):
  - Session-detail `GET .../sessions/:runner/:id` count is **exactly 1** (initial page-load snapshot only).
  - Exactly **one** `GET .../events/stream` request started.
  - **Zero** stream requests aborted or cancelled (`net::ERR_ABORTED` / client abort).
  - No ~5s interval detail GET pattern (meta poll must be absent).

## Side Effects

- New `grok-tty` session created via API POST (`runner: grok-tty`).
- Background `agent-run web` serves SSE and session detail endpoints.

## Exit Code

- Playwright process exits 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "no-detail-poll" {
		t.Fatalf("expected layout no-detail-poll, got %q", req.Layout)
	}
}
```