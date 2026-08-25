# Scenario

**Bug**: Marcus menu shows `Codex: Error: run: session id "codex-status-usage" already in use`
when a reclaimable leftover tty-watch claim/registry still holds the fixed usage session id.
`FetchStatus` returns that reserve error immediately; headless `Run` already reclaim-once-and-retries.

```
plant reclaimable hold on codex-status-usage  # .claim with alive PID (no live PTY)
usage.Fetch(Codex) -> FetchStatus -> ReserveCustomSessionID
# want: reclaim zombie + fetch snapshot (like headless)
# today: Error already in use  (Marcus menu caches it)
```

## Preconditions

- Provider is Codex; Mode=fetch (default).
- Isolated `TTY_WATCH_HOME` under tempdir.
- Default fake Codex TUI (same as `fetch-codex-mock`).
- Leaves set `PlantAliveSessionClaim` (and optional other stale shapes later).

## Steps

1. Set Provider=Codex + default fake status command + SessionID=`codex-status-usage`.
2. Leaf enables `PlantAliveSessionClaim`.
3. Run plants claim then calls `usage.FetchWithOptions`.
4. Assert expects a successful Snapshot (RED until FetchStatus reclaims).

## Context

- Claim file path: `{TTY_WATCH_HOME}/registry/.codex-status-usage.claim`
- Alive claim PID is kept by `pruneStaleSessionID`; `shouldReclaimZombieForReserve`
  treats claim-only holds as reclaimable (headless path).

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Provider = usage.Codex
	req.ShowStatusCommand = fakeCodexTUIDefault()
	req.SessionID = "codex-status-usage"
	return nil
}
```
