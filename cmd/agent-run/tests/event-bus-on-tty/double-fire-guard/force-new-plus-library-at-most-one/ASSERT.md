## Expected

- `PublishCount == 1` (at most one — and when URL set, exactly one — per open).
- Captured type/source are `agent.tty.started` / `agent-run`.
- Payload session_id matches leaf SessionID.

## Side Effects

- Single inject publish despite both ForceNew and library call sites.

## Errors

- None.

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.PublishCount != 1 {
		t.Fatalf("ForceNew+library must publish at most once per open; PublishCount=%d (want 1)",
			resp.PublishCount)
	}
	got, ok := req.Capture.Last()
	if !ok {
		t.Fatal("missing captured publish")
	}
	if got.Type != wireTypeAgentTTYStarted {
		t.Fatalf("type: got %q, want %q", got.Type, wireTypeAgentTTYStarted)
	}
	if got.Source != wireSourceAgentRun {
		t.Fatalf("source: got %q, want %q", got.Source, wireSourceAgentRun)
	}
	var payload map[string]any
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if payload["session_id"] != req.SessionID {
		t.Fatalf("payload.session_id: got %v want %q", payload["session_id"], req.SessionID)
	}
}
```
