---
label: e2e
---

## Expected

- Session detail JSON includes assistant content with `WEB_E2E_GROK_HARNESS_MARKER`.
- Status is `finished`, or `running` under keep-tty once the marker has streamed.
- Runner is `grok-tty` (not `fake-codex`).

## Side Effects

- Session dir under `AGENT_RUN_HOME/sessions/grok-tty/<session_id>/`.
- Grok home tree under `--grok-home` contains `updates.jsonl` for mock UUID.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("session detail status=%d body=%s", resp.HTTPStatus, resp.HTTPBody)
	}
	if !strings.Contains(resp.HTTPBody, harnessSmokeMarker) {
		t.Fatalf("expected harness marker %q in detail/events path; body=%s", harnessSmokeMarker, resp.HTTPBody)
	}
	finished := strings.Contains(resp.HTTPBody, `"status":"finished"`) || strings.Contains(resp.HTTPBody, `"status": "finished"`)
	running := strings.Contains(resp.HTTPBody, `"status":"running"`) || strings.Contains(resp.HTTPBody, `"status": "running"`)
	if !finished && !running {
		t.Fatalf("expected finished or running (keep-tty) status, got: %s", resp.HTTPBody)
	}
	if !strings.Contains(resp.HTTPBody, "grok-tty") {
		t.Fatalf("expected grok-tty runner in detail: %s", resp.HTTPBody)
	}
}
```