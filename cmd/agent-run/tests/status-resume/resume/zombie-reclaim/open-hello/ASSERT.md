## Expected

- Must **not** fail with `session id "…" already in use` / `already in use`.
- Resume gate is open (zombie is exited); reclaim frees the registry id so
  `ReserveCustomSessionID` can reuse the same terminal id.
- Prefer exit 0 (instant attach + fake TUI completes open path).
- On success, `meta.terminal_session_id` is still the original id (primary
  reclaim reuse) or a non-empty fallback id — never blocked by the old zombie.

## Side Effects

- Zombie registry entry is reclaimed (torn down / removed) so the same id can
  be reserved, or a new terminal id is allocated without "already in use".

## Exit Code

0 (preferred after reclaim)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("exec error: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := resp.Stderr + "\n" + resp.Stdout
	low := strings.ToLower(combined)
	if strings.Contains(low, "already in use") {
		t.Fatalf("resume --open with zombie registry must reclaim/reuse terminal id, not fail with already-in-use:\n%s", combined)
	}
	// Gate must have opened (zombie = exited). Live-not-exited would be a fixture bug.
	if strings.Contains(low, "cannot resume") && (strings.Contains(low, "not exited") || strings.Contains(low, "still active")) {
		t.Fatalf("fixture should be zombie-exited (resume ready), not live gate deny:\n%s", combined)
	}
	assertSuccess(t, resp)
	if resp.MetaAfter != nil {
		termID, _ := resp.MetaAfter["terminal_session_id"].(string)
		termID = strings.TrimSpace(termID)
		if termID == "" {
			t.Fatalf("after successful resume --open, meta.terminal_session_id must be non-empty; meta=%v", resp.MetaAfter)
		}
		// Primary (A): reuse same id. Fallback (B): new id is allowed if reclaim
		// could not free the old entry — still not "already in use".
		if termID != req.SessionID && termID != req.TerminalSessionID {
			t.Logf("terminal_session_id changed to %q (fallback B ok if reclaim failed); original session=%q", termID, req.SessionID)
		}
	}
}
```
