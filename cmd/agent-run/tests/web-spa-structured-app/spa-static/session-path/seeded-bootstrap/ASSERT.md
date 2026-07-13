## Expected

- HTTP **200**.
- Body contains `#root`.
- Body includes `<script id="agent-run-session-bootstrap" …>` (or equivalent id attribute).
- Bootstrap payload includes the seeded `session_id` (`spa-seed-bootstrap`).

## Side Effects

- Session files under `AGENT_RUN_HOME/sessions/spa-seed-bootstrap/`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	if !htmlHasRootMount(resp.HTTPBody) {
		t.Fatalf("missing #root")
	}
	if !htmlHasSessionBootstrap(resp.HTTPBody) {
		t.Fatalf("expected agent-run-session-bootstrap injection for seeded session; body=%q", resp.HTTPBody)
	}
	if !bootstrapContainsSessionID(resp.HTTPBody, req.SessionID) {
		t.Fatalf("bootstrap missing session_id %q; body=%q", req.SessionID, resp.HTTPBody)
	}
}
```
