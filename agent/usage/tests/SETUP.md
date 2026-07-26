# Scenario

**Feature**: unified usage.Fetch facade returns normalized snapshots via env fake hooks;
Codex FetchStatus must release the registry exclusive flock before long TUI work

```
# isolated TTY_WATCH_HOME for codex; GROK/CODEX_SHOW_*_COMMAND fake TUIs
leaf Setup sets provider + fake hook -> Run calls usage.Fetch(ctx, provider)
doctest <- Snapshot provider, usage %, reset, credits fields

# lock-lifetime leaves
FetchStatus (blocking fake TUI) mid-flight
  -> concurrent LOCK_NB / ReserveCustomSessionID on same home
doctest <- lock free while fetch still running; same-id still in use
```

## Preconditions

- `agent/usage` package exports `Fetch`, `ProviderID`, and `Snapshot`.
- `agent/grok/tty` honors `GROK_SHOW_USAGE_COMMAND`.
- `agent/codex/tty` honors `CODEX_SHOW_STATUS_COMMAND` and uses in-process ttywatch
  (no `exec tty-watch` subprocess).
- Codex leaves use isolated `TTY_WATCH_HOME` under `t.TempDir()` for **test isolation only** —
  not as product guidance to run Codex under a private home.
- After successful `ReserveCustomSessionID`, `FetchStatus` must release the registry flock
  **before** `StartInProcess` / prompt-status waits (must not `defer release()` across the whole fetch).

## Steps

1. Root `Setup` creates `req.TTYWatchHome` for codex provider tests.
2. Leaf `Setup` sets `req.Provider`, Mode, and the appropriate fake command hook.
3. `Run` either calls `usage.Fetch` (Mode=fetch) or starts a long fetch and probes the lock
   (Mode=lock-during-fetch).
4. Leaf `Assert` checks Snapshot fields or mid-fetch lock observations.

## Context

- Fake TUI scripts mirror `script/grok/show-usage/tests` and `script/codex/show-status/tests` fixtures.
- Grok snapshot uses `UsagePercent` for weekly limit; `CreditsUsed`/`CreditsTotal` stay empty.
- Codex snapshot fills all credit fields from `/status` parse.
- Blocking Codex fake never reaches a writable prompt so `FetchStatus` stays in wait while the
  concurrent probe runs; harness cancels the context after probing.
- Isolated home is per-leaf tempdir only; product still uses the shared default home semantics.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TTYWatchHome == "" {
		req.TTYWatchHome = filepath.Join(t.TempDir(), ".tty-watch")
	}
	return nil
}

// fakeGrokTUIDefault mimics grok TUI: prompt, read command, print fixture usage lines.
func fakeGrokTUIDefault() string {
	return `sh -c 'printf "Grok › "; read -r cmd; printf "Weekly limit: 1%%\nNext reset: July 9, 16:55 PT\n› "'`
}

// fakeCodexTUIDefault mimics codex TUI: prompt, read /status, print canonical fixture status box.
func fakeCodexTUIDefault() string {
	return `sh -c 'printf "Codex › "; read -r cmd; printf "Monthly credit limit: 42%% left (resets 08:00 on 1 Aug)\n6,519 of 11,250 credits used\n› "'`
}

// fakeCodexTUIBlocking never becomes idle / never prints status fields so FetchStatus
// remains in StartInProcess / wait-for-prompt while concurrent lock probes run.
func fakeCodexTUIBlocking() string {
	return `sh -c 'printf "codex-status-block\n"; sleep 120'`
}

func assertSnapshotField(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", field, got, want)
	}
}
```
