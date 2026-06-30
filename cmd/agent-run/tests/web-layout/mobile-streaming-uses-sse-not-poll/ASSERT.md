---
label: chromium, slow
explanation: Live fake-codex run; 8s network monitor window
---

## Expected

- Playwright exit code **0**.
- Viewport 390×844; no horizontal document scroll.
- During the 8s monitoring window after opening a live streaming session:
  - At least one `GET .../events/stream` request.
  - Session-detail `GET .../sessions/:runner/:id` count is **≤ 3** (initial load + at most two meta refresh).
  - No 250ms poll storm (detail GET count must stay bounded).

## Side Effects

- New `fake-codex` session created via API POST.
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
	if req.Layout != "sse-transport" {
		t.Fatalf("expected layout sse-transport, got %q", req.Layout)
	}
}
```