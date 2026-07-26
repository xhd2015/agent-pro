# Scenario

**Bug**: Codex `FetchStatus` holds the registry exclusive flock for the entire fetch
(`defer release()` after `ReserveCustomSessionID`), blocking concurrent `tty-watch` /
reserve on the same home while StartInProcess and status waits run.

```
# required lock lifetime
FetchStatus -> ReserveCustomSessionID -> release()   # flock free
            -> StartInProcess / wait-prompt /status  # long work without flock

# concurrent peer on same isolated TTY_WATCH_HOME (test isolation only)
peer -> LOCK_NB registry/.lock  OR  ReserveCustomSessionID(other-id)
doctest <- peer succeeds while fetch still in progress
```

## Preconditions

- Provider is Codex.
- `Mode=lock-during-fetch`.
- Fake TUI blocks (no prompt / no status) so the fetch stays mid-flight.
- `TTY_WATCH_HOME` is an isolated temp dir (isolation only; not product alternate-home).

## Steps

1. Set Provider=Codex, Mode=lock-during-fetch.
2. Install blocking `CODEX_SHOW_STATUS_COMMAND` fake.
3. Default session id `codex-status-usage`; probe other id `probe-other-session`.
4. Shorten provider timeout env so hung TUIs cannot linger after cancel.
5. Leaves refine SameIDProbe and assert mid-fetch lock / same-id behavior.

## Context

- Sound fix (implementer later): call `release()` immediately after successful reserve;
  keep `defer session.Kill()`; honor context cancel/timeout.
- Does **not** require running Codex under another product `TTY_WATCH_HOME`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Provider = usage.Codex
	req.Mode = "lock-during-fetch"
	req.ShowStatusCommand = fakeCodexTUIBlocking()
	req.SessionID = "codex-status-usage"
	req.ProbeSessionID = "probe-other-session"
	// Keep FetchStatus deadline bounded; harness also cancels after probe.
	req.TimeoutSeconds = "20"
	return nil
}
```
