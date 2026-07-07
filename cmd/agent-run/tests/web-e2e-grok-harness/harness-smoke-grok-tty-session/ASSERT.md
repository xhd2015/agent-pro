## Expected

- POST `grok-tty` session reaches `status=finished`.
- Session detail JSON includes assistant content with `WEB_E2E_GROK_HARNESS_MARKER`.
- Runner is `grok-tty` (not `fake-codex`).

## Side Effects

- Session dir under `AGENT_RUN_HOME/sessions/grok-tty/<session_id>/`.
- Grok home tree under `--grok-home` contains `updates.jsonl` for mock UUID.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("session detail status=%d body=%s", resp.HTTPStatus, resp.HTTPBody)
	}
	if !strings.Contains(resp.HTTPBody, `"status":"finished"`) && !strings.Contains(resp.HTTPBody, `"status": "finished"`) {
		t.Fatalf("expected finished status, got: %s", resp.HTTPBody)
	}
	if !strings.Contains(resp.HTTPBody, "grok-tty") {
		t.Fatalf("expected grok-tty runner in detail: %s", resp.HTTPBody)
	}
	if !strings.Contains(resp.HTTPBody, harnessSmokeMarker) {
		t.Fatalf("expected harness marker %q in detail/events path; body=%s", harnessSmokeMarker, resp.HTTPBody)
	}
}
```