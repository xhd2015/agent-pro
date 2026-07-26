---
label: e2e
---

## Expected

- Registry file exists at `AGENT_RUN_HOME/grok-tty-registry/<session-id>.json` while run is active.
- JSON contains `session_id`, `listen_addr` (`127.0.0.1:<port>`), `pid` > 0, non-empty `created_at`.
- `listen_addr` accepts TCP connections (adhoc ptywrap server listening).

## Side Effects

- Background `agent-run run` started during Setup; killed on test cleanup.

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
	if resp.RegistryEntry == nil {
		t.Fatal("expected registry entry while run is active")
	}
	entry := resp.RegistryEntry
	if entry.SessionID != req.GrokTTYSessionID {
		t.Fatalf("session_id: got %q want %q", entry.SessionID, req.GrokTTYSessionID)
	}
	if !strings.HasPrefix(entry.ListenAddr, "127.0.0.1:") {
		t.Fatalf("listen_addr should be 127.0.0.1:<port>, got %q", entry.ListenAddr)
	}
	if entry.PID <= 0 {
		t.Fatalf("expected positive pid, got %d", entry.PID)
	}
	if strings.TrimSpace(entry.CreatedAt) == "" {
		t.Fatal("expected non-empty created_at")
	}
	if !portOpen(entry.ListenAddr) {
		t.Fatalf("adhoc server not listening on %s", entry.ListenAddr)
	}
	status, err := httpGetHealth(entry.ListenAddr)
	if err != nil {
		t.Fatalf("ptywrap health probe %s: %v", entry.ListenAddr, err)
	}
	if status != 200 {
		t.Fatalf("GET /api/terminal/sessions status=%d want 200", status)
	}
}
```