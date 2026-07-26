---
label: e2e
---

## Expected

- Exit code 0.
- Stderr registry id has base `hello-world` and numeric collision suffix `-N` (N ≥ 1).
- Storage session id equals stderr id.
- Registry file exists for that same id.
- `meta.terminal_session_id` equals the same id.

## Exit Code

0

```go
import (
	"os"
	"strconv"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	stderrID, ok := parseGrokTTYSessionID(resp.Stderr)
	if !ok {
		t.Fatalf("missing grok-tty session id on stderr:\n%s", resp.Stderr)
	}
	base, _, suf, ok := splitAutoSessionID(stderrID)
	if !ok || base != "hello-world" {
		t.Fatalf("stderr id %q want hello-world-… with collision suffix", stderrID)
	}
	if suf == "" {
		t.Fatalf("expected numeric collision suffix on id %q", stderrID)
	}
	n, convErr := strconv.Atoi(suf)
	if convErr != nil || n < 1 {
		t.Fatalf("collision suffix %q invalid on id %q", suf, stderrID)
	}

	// Storage may contain seed dirs; the live id must exist and match stderr.
	if _, err := os.Stat(sessionDir(req.Home, "grok-tty", stderrID)); err != nil {
		t.Fatalf("storage dir for %q missing: %v", stderrID, err)
	}
	meta := readSessionMeta(t, req.Home, "grok-tty", stderrID)
	if meta.SessionID != stderrID {
		t.Fatalf("meta.session_id = %q, want %q", meta.SessionID, stderrID)
	}
	if meta.TerminalSessionID != stderrID {
		t.Fatalf("meta.terminal_session_id = %q, want %q", meta.TerminalSessionID, stderrID)
	}
	regPath := grokTTYRegistryPath(req.Home, stderrID)
	if _, err := os.Stat(regPath); err != nil {
		t.Fatalf("registry missing for %q: %v", stderrID, err)
	}
}
```
