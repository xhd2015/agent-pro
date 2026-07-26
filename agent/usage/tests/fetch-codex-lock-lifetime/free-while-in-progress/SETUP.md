# Scenario

**Bug**: while a long-running Codex usage fetch is in progress, the registry exclusive
flock must already be released so another reserve / lock acquire on the same home can succeed

```
# mid-fetch after reserve (+ optional StartInProcess)
FetchStatus holds only claim/session, not registry .lock
peer LOCK_NB(.lock) -> ok
peer ReserveCustomSessionID(probe-other-session) -> ok
harness cancel -> session.Kill -> registry entry gone
```

## Preconditions

- Inherits Codex blocking fake and Mode=lock-during-fetch from parent.
- SameIDProbe left false (sibling leaf covers same-id).

## Steps

1. Confirm Mode=lock-during-fetch and blocking fake.
2. Run starts `usage.Fetch(Codex)` in a goroutine.
3. After claim or registry entry appears, probe LOCK_NB and other-id reserve.
4. Cancel fetch; observe cleanup.

## Context

- RED against current code: `defer release()` keeps flock held until FetchStatus returns,
  so LOCK_NB fails with EWOULDBLOCK / EAGAIN and other-id reserve hits lock-busy timeout.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Parent already set Mode, Provider, blocking command, session ids.
	req.SameIDProbe = false
	return nil
}
```
