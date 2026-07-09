## Expected

- Exit code 0.
- Stderr contains `grok-tty: my-task` (not `session-N`).
- Storage `sessions/grok-tty/my-task/` exists with `meta.session_id == my-task`.
- `meta.terminal_session_id == my-task`.
- Registry `grok-tty-registry/my-task.json` exists (`--keep-tty`).

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	const want = "my-task"
	stderrID, ok := parseGrokTTYSessionID(resp.Stderr)
	if !ok {
		t.Fatalf("missing grok-tty session id on stderr:\n%s", resp.Stderr)
	}
	if stderrID != want {
		t.Fatalf("stderr id = %q, want %q\nstderr:\n%s", stderrID, want, resp.Stderr)
	}
	// Ensure default session-N numbering was not used as the primary id line.
	if strings.Contains(resp.Stderr, "grok-tty: session-") && stderrID != want {
		t.Fatalf("unexpected default session-N id while --session %s set:\n%s", want, resp.Stderr)
	}

	dir := sessionDir(req.Home, "grok-tty", want)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("storage dir missing %s: %v\nids=%v", dir, err, listSessionIDs(t, req.Home, "grok-tty"))
	}
	meta := readSessionMeta(t, req.Home, "grok-tty", want)
	if meta.SessionID != want {
		t.Fatalf("meta.session_id = %q, want %q", meta.SessionID, want)
	}
	if meta.TerminalSessionID != want {
		t.Fatalf("meta.terminal_session_id = %q, want %q", meta.TerminalSessionID, want)
	}
	regPath := grokTTYRegistryPath(req.Home, want)
	if _, err := os.Stat(regPath); err != nil {
		t.Fatalf("registry missing at %s: %v\nstderr:\n%s", regPath, err, resp.Stderr)
	}
}
```
