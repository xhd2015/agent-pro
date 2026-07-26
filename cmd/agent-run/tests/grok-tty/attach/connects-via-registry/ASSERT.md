---
label: e2e
---

## Expected

- Registry entry found for parsed session id.
- Attach probe succeeds: WebSocket handshake to hidden `listen_addr` with snapshot.
- Probe does not fail with registry lookup or connection errors.

## Side Effects

- Background `agent-run run` started during Setup; killed on test cleanup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.RegistryEntry == nil {
		t.Fatal("expected registry entry for attach probe")
	}
	if resp.RegistryEntry.SessionID != req.GrokTTYSessionID {
		t.Fatalf("registry session_id %q != stderr id %q", resp.RegistryEntry.SessionID, req.GrokTTYSessionID)
	}
	if !resp.AttachProbeOK {
		t.Fatalf("attach probe failed: %s", resp.AttachProbeErr)
	}
}
```