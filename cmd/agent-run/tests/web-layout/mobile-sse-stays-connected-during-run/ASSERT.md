---
label: e2e, chromium, slow
explanation: Seeded running session with no live event writer; 8s SSE idle-gap monitor
---

## Expected

- Playwright exit code **0**.
- Viewport 390×844; no horizontal document scroll.
- During the 8s monitoring window after opening a seeded `running` session (user message only, no assistant reply, no live agent writing events):
  - Exactly **one** `GET .../events/stream` request started.
  - **Zero** stream requests aborted or cancelled (`net::ERR_ABORTED` / client abort) — the 2s idle timer must not kill the connection.
  - Session-detail `GET .../sessions/:runner/:id` count is **≤ 3** (no poll fallback storm).

## Side Effects

- Seeded session files under `AGENT_RUN_HOME/sessions/fake-opencode/sse-persist/`.
- Background `agent-run web` serves SSE tail against static `events.jsonl` (no new lines appended).

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
	if req.Layout != "sse-persistence" {
		t.Fatalf("expected layout sse-persistence, got %q", req.Layout)
	}
}
```