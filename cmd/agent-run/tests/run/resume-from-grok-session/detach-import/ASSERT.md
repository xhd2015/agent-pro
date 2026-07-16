## Expected

- Exit code 0 (parent returns after detach registry; not after full hold sleep).
- Stdout contains labeled `session-id:` and `terminal-id:` lines (detach contract).
- Meta for the agent-run session has `runner_session_id` equal to the Grok UUID.
- Soft: registry file under `grok-tty-registry/<terminal-id>.json` may exist.

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
		t.Fatalf("run failed (expect detach to return before hold timeout): %v\nstdout:\n%s\nstderr:\n%s",
			err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)
	sid, tid := assertDetachIDsOnStdout(t, resp)

	// Prefer fixed --session-id when provided.
	if req.SessionID != "" && sid != req.SessionID {
		// Product may still print the same id; allow only exact match preference.
		// If labels show a different id, still require meta bind under req.SessionID
		// when that path exists.
		t.Logf("detach session-id=%q (cli --session-id=%q)", sid, req.SessionID)
	}

	metaID := req.SessionID
	if metaID == "" {
		metaID = sid
	}
	meta := readSessionMetaFile(t, req.Home, metaID)
	gotRSID, _ := meta["runner_session_id"].(string)
	if gotRSID != req.GrokSessionID {
		t.Fatalf("meta.runner_session_id = %q, want %q; meta=%v", gotRSID, req.GrokSessionID, meta)
	}
	if r, _ := meta["runner"].(string); r != "" && r != "grok-tty" {
		t.Fatalf("meta.runner = %q, want grok-tty", r)
	}

	// Soft registry presence (keep-alive).
	reg := registryPath(req.Home, "grok-tty", tid)
	if _, err := os.Stat(reg); err != nil {
		// try session-id as registry key
		alt := registryPath(req.Home, "grok-tty", sid)
		if _, e2 := os.Stat(alt); e2 != nil {
			t.Logf("note: registry not found at %s or %s (soft): %v", reg, alt, err)
		}
	}
	_ = strings.TrimSpace(tid)
}
```
