## Expected

- Reserve completed: `ClaimOrRegistrySeen` is true.
- Mid-fetch, registry flock is free (`LockAcquiredDuring=true`) so peers are not stuck on flock.
- Concurrent `ReserveCustomSessionID` for the **same** usage session id fails
  (`SameIDReserveSucceeded=false`).
- Failure text indicates the id is already in use (substring `already in use`), **not**
  registry lock-busy / flock timeout.
- After cancel, fetch returns and registry entry is cleaned up.

## Side Effects

- Live claim/session for `codex-status-usage` remains until Kill; same-id re-reserve rejected.
- Isolated temp home only.

## Errors

- Same-id reserve succeeding means claim/live check is broken.
- Same-id error that only mentions lock busy / timed out waiting for exclusive flock means
  release still spans the long fetch (current bug).
- Flock still held (`LockAcquiredDuring=false`) is also RED for early-release.

## Exit Code

N/A (in-process package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if !resp.ClaimOrRegistrySeen {
		t.Fatalf("expected claim or registry entry before same-id probe; fetch_err=%q", resp.FetchErr)
	}

	// Early release must already have happened so we observe session-in-use, not lock-busy.
	if !resp.LockAcquiredDuring {
		t.Fatalf("expected registry flock free during fetch so same-id check can run; lock_err=%q same_id_err=%q",
			resp.LockAcquireErr, resp.SameIDReserveErr)
	}
	if resp.SameIDReserveSucceeded {
		t.Fatalf("expected ReserveCustomSessionID(%q) to fail while first session is live", req.SessionID)
	}
	msg := strings.ToLower(resp.SameIDReserveErr)
	if msg == "" {
		t.Fatal("expected non-empty SameIDReserveErr")
	}
	if strings.Contains(msg, "lock busy") || strings.Contains(msg, "exclusive flock") ||
		strings.Contains(msg, "timed out") && strings.Contains(msg, "lock") {
		t.Fatalf("same-id reserve must fail with already-in-use, not lock-busy; got %q", resp.SameIDReserveErr)
	}
	if !strings.Contains(msg, "already in use") {
		t.Fatalf("same-id reserve error %q, want substring %q", resp.SameIDReserveErr, "already in use")
	}

	if !resp.FetchCompleted {
		t.Fatalf("expected fetch to return after cancel; fetch_err=%q", resp.FetchErr)
	}
	if !resp.RegistryGoneAfter {
		t.Fatalf("expected registry entry for %q gone after cancel/Kill", req.SessionID)
	}
}
```
