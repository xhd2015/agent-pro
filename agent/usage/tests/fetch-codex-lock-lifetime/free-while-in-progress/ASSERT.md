## Expected

- Reserve completed before probes: `ClaimOrRegistrySeen` is true.
- While the long fetch is still running (before cancel), non-blocking exclusive flock on
  `{TTY_WATCH_HOME}/registry/.lock` **succeeds** (`LockAcquiredDuring=true`).
- Concurrent `ReserveCustomSessionID` for a **different** session id on the same home
  **succeeds** (`SecondReserveOK=true`) — proves peers are not blocked by usage fetch.
- After cancel, fetch returns (`FetchCompleted=true`) with a non-empty error (cancel/timeout).
- Registry JSON for the usage session id is gone after cleanup (`RegistryGoneAfter=true`).

## Side Effects

- Ephemeral session for `codex-status-usage` is killed on cancel.
- Probe session claim for `probe-other-session` is removed by the harness after release.
- Isolated temp `TTY_WATCH_HOME` only — not a product “alternate home” requirement.

## Errors

- `LockAcquiredDuring=false` with lock busy / EWOULDBLOCK means flock was still held across
  StartInProcess/wait (current bug: `defer release()` for whole fetch).
- `SecondReserveOK=false` with registry lock-busy diagnostics is the same failure mode.
- Fetch finishing before any claim/registry appears means the blocking fake or reserve failed.

## Exit Code

N/A (in-process package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if !resp.ClaimOrRegistrySeen {
		t.Fatalf("expected claim or registry entry for session %q before mid-fetch probe; fetch_err=%q",
			req.SessionID, resp.FetchErr)
	}

	// Primary RED assertion: flock must not be held for the duration of StartInProcess/wait.
	if !resp.LockAcquiredDuring {
		t.Fatalf("expected non-blocking exclusive flock on registry/.lock to succeed while FetchStatus is in progress; lock_err=%q second_reserve_err=%q (current code defers release() across whole fetch)",
			resp.LockAcquireErr, resp.SecondReserveErr)
	}
	if !resp.SecondReserveOK {
		t.Fatalf("expected ReserveCustomSessionID(%q) to succeed on same home during long fetch; err=%q",
			req.ProbeSessionID, resp.SecondReserveErr)
	}

	if !resp.FetchCompleted {
		t.Fatalf("expected fetch to return after cancel; fetch_err=%q", resp.FetchErr)
	}
	if strings.TrimSpace(resp.FetchErr) == "" {
		t.Fatal("expected fetch to end with cancel/timeout error after blocking fake + cancel")
	}
	if !resp.RegistryGoneAfter {
		t.Fatalf("expected registry entry for %q to be gone after cancel/Kill", req.SessionID)
	}
}
```
