---
label: codex
explanation: Requires real codex CLI on PATH; for design verification and debugging.
---

## Expected

- Background run publishes `codex-tty: session-N` on stderr.
- Attach probe succeeds while run is still active (registry + WS handshake).
- Attach snapshot or scrollback shows codex TUI output (non-empty terminal bytes).

## Side Effects

- Background `agent-run run` started during Setup; killed on test cleanup.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil && !resp.AttachProbeOK {
		t.Fatal(err)
	}
	if req.CodexTTYSessionID == "" {
		t.Fatal("expected session id from background run stderr")
	}
	if resp.RegistryEntry == nil {
		t.Fatal("expected registry entry while real-codex run is active")
	}
	if !resp.AttachProbeOK {
		t.Fatalf("attach while running failed: %s", resp.AttachProbeErr)
	}
	combined := resp.Stdout + resp.Stderr + resp.BackgroundStdout + resp.BackgroundStderr
	if strings.TrimSpace(combined) == "" {
		t.Fatal("expected visible codex output in run or attach streams")
	}
}
```