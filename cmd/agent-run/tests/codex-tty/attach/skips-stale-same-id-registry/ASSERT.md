## Expected

- `agent-run attach session-1` does not stop at the stale grok-tty registry entry.
- The CLI attach attempt reaches the live codex-tty session and is treated as connected.
- Stderr must not report connection refused for the stale grok-tty port.

## Side Effects

- Background `agent-run run` started during Setup; killed on test cleanup.
- A stale `grok-tty-registry/session-1.json` exists in the isolated test home.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if req.CodexTTYSessionID != "session-1" {
		t.Fatalf("expected same-id collision on session-1, got %q", req.CodexTTYSessionID)
	}
	if resp.RegistryEntry == nil {
		t.Fatal("expected live codex-tty registry entry to remain available")
	}
	if resp.RegistryEntry.SessionID != "session-1" {
		t.Fatalf("live registry session_id = %q, want session-1", resp.RegistryEntry.SessionID)
	}
	if !resp.AttachProbeOK {
		t.Fatalf("attach CLI did not reach live codex-tty session; exit=%d stderr=%q probe=%q", resp.ExitCode, resp.Stderr, resp.AttachProbeErr)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.AttachProbeErr)
	if strings.Contains(combined, "connection refused") || strings.Contains(combined, "connect:") {
		t.Fatalf("attach stopped at stale registry entry instead of skipping it:\n%s", combined)
	}
}
```
