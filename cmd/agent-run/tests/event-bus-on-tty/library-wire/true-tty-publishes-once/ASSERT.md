## Expected

- No API error from AutoSendOrResume.
- `RunCalls == 1` (ModeRun dispatch).
- `PublishCount == 1` (one HTTP publish via WireOnTTYStarted, not ForceNew label).
- Captured event type/source are `agent.tty.started` / `agent-run`.
- Payload includes session_id / runner / workspace from the leaf.

## Side Effects

- One inject publish from library OnTTYStarted wire.

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
	if resp.APIErrString != "" {
		t.Fatalf("AutoSendOrResume: %s", resp.APIErrString)
	}
	if resp.RunCalls != 1 {
		t.Fatalf("RunCalls: got %d, want 1", resp.RunCalls)
	}
	if resp.PublishCount != 1 {
		t.Fatalf("true-TTY wire must publish once; PublishCount=%d", resp.PublishCount)
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
	if payload["runner"] != req.Runner {
		t.Fatalf("payload.runner: got %v want %q", payload["runner"], req.Runner)
	}
	if payload["workspace"] != req.Workspace {
		t.Fatalf("payload.workspace: got %v want %q", payload["workspace"], req.Workspace)
	}
}
```
