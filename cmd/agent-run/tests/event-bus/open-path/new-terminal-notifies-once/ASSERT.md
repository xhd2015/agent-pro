## Expected

- `PublishCount` == 1 (exactly one NotifyTTYStarted publish after successful open).
- Captured event type/source are `agent.tty.started` / `agent-run`.
- Payload includes session_id / runner / workspace from the leaf.
- `ResultArgs` include `--event-bus-url` and `--event-bus-token` with leaf values
  (ForceNew follow-up carries flags so the child can publish too).

## Side Effects

- One inject publish; follow-up argv extended.

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
		t.Fatalf("new-terminal success must publish once; PublishCount=%d", resp.PublishCount)
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

	if !hasFlagValue(resp.ResultArgs, "--event-bus-url", req.EventBusURL) {
		t.Fatalf("follow-up must include --event-bus-url; args=%#v", resp.ResultArgs)
	}
	if !hasFlagValue(resp.ResultArgs, "--event-bus-token", req.EventBusToken) {
		t.Fatalf("follow-up must include --event-bus-token; args=%#v", resp.ResultArgs)
	}
}

// hasFlagValue accepts either ["--flag", "value"] or ["--flag=value"].
func hasFlagValue(args []string, flag, value string) bool {
	eq := flag + "=" + value
	for i, a := range args {
		if a == eq {
			return true
		}
		if a == flag && i+1 < len(args) && args[i+1] == value {
			return true
		}
	}
	return false
}
```
