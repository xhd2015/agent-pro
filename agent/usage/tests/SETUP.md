# Scenario

**Feature**: unified usage.Fetch facade returns normalized snapshots via env fake hooks

```
# isolated TTY_WATCH_HOME for codex; GROK/CODEX_SHOW_*_COMMAND fake TUIs
leaf Setup sets provider + fake hook -> Run calls usage.Fetch(ctx, provider)
doctest <- Snapshot provider, usage %, reset, credits fields
```

## Preconditions

- `agent/usage` package exports `Fetch`, `ProviderID`, and `Snapshot` (added by implementer).
- `agent/grok/tty` honors `GROK_SHOW_USAGE_COMMAND`.
- `agent/codex/tty` honors `CODEX_SHOW_STATUS_COMMAND` and uses in-process ttywatch
  (no `exec tty-watch` subprocess).
- Codex leaves use isolated `TTY_WATCH_HOME` under `t.TempDir()`.

## Steps

1. Root `Setup` creates `req.TTYWatchHome` for codex provider tests.
2. Leaf `Setup` sets `req.Provider` and the appropriate fake command hook.
3. `Run` applies env overrides and calls `usage.Fetch`.
4. Leaf `Assert` checks normalized `Snapshot` fields.

## Context

- Fake TUI scripts mirror `script/grok/show-usage/tests` and `script/codex/show-status/tests` fixtures.
- Grok snapshot uses `UsagePercent` for weekly limit; `CreditsUsed`/`CreditsTotal` stay empty.
- Codex snapshot fills all credit fields from `/status` parse.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
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

func assertSnapshotField(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", field, got, want)
	}
}
```