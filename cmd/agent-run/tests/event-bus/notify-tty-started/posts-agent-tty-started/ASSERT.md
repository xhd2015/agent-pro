## Expected

- Exactly one publish recorded.
- Event `type` == `agent.tty.started` (eventbus.TypeAgentTTYStarted).
- Event `source` == `agent-run` (eventbus.SourceAgentRun).
- Payload JSON object has:
  - `session_id` == req.SessionID
  - `runner` == req.Runner
  - `workspace` == req.Workspace
- WarnOutput empty (success path).

## Side Effects

- One inject PublishHook call.

## Errors

- Run `err` is nil.

```go
import (
	"encoding/json"
	"strings"
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
		t.Fatalf("PublishCount: got %d, want 1", resp.PublishCount)
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
		t.Fatalf("payload JSON: %v\n%s", err, got.Payload)
	}
	for key, want := range map[string]string{
		"session_id": req.SessionID,
		"runner":     req.Runner,
		"workspace":  req.Workspace,
	} {
		gotV, ok := payload[key]
		if !ok {
			t.Fatalf("payload missing %q; payload=%v", key, payload)
		}
		if s, _ := gotV.(string); s != want {
			t.Fatalf("payload[%q]: got %v, want %q", key, gotV, want)
		}
	}
	if strings.TrimSpace(resp.WarnOutput) != "" {
		t.Fatalf("success path must not warn; WarnOutput=%q", resp.WarnOutput)
	}
}
```
