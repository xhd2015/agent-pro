# Scenario

**Feature**: ephemeral session is removed from tty-watch registry after fetch

```
# successful fetch -> tty-watch kill -> registry entry for session id is gone
codex-show-status -> tty-watch run/send/snapshot/kill -> registry pruned
```

## Preconditions

- `TTY_WATCH_HOME` is isolated under `req.TempDir`.
- Session id is `codex-status-usage` (default `CODEX_SHOW_STATUS_SESSION_ID`).

## Steps

1. Inherit default fake TUI from `success/SETUP.md`.
2. Run CLI to completion.
3. Assert `codex-status-usage` is absent from `$TTY_WATCH_HOME/registry/`.

## Context

- Verifies cleanup runs on success paths, not only error paths.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ShowStatusCommand = fakeTUIDefault()
	req.SessionID = "codex-status-usage"
	return nil
}
```