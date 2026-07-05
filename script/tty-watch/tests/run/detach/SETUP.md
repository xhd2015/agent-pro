# Scenario

**Feature**: `tty-watch run --detach` starts a PTY session without writer attach and exits immediately

```
# detach parent: reserve id, spawn serve child, skip attachWriter, print session-id, exit 0
harness -> tty-watch run --detach <cmd> -> stdout: session-id: <id>\n -> exit 0 promptly
# serve child + registry survive; list/watch/attach/kill still work
tty-watch list -> session live after detach parent exited
# --detach and --headless are mutually exclusive (error before start)
```

## Preconditions

- Isolated `TTY_WATCH_HOME` from root Setup.
- Detach runs use **pipe stdin** (not PTY attach); harness waits for parent exit 0.
- `buildDetachRunArgv` always inserts `--detach` before optional `--session-id` and `<command>`.
- Harness kills detached sessions via `kill <id>` in test cleanup.

## Steps

1. Leaf sets `req.Phase` to a `run-detach-*` key and optional `RunCommand` / `CustomSessionID`.
2. Grouping helpers run detach via `runCLISeparate`, parse `session-id: <id>\n`, then probe list or watch.
3. `Run` (root `DOCTEST.md`) dispatches `run-detach-*` phases to helpers below before `ttywatchtest.Run`.

## Context

- Detach parent exits 0 immediately; serve child keeps registry entry alive.
- No SIGINT handler or waiting logs on detach parent.
- `--detach` + `--headless` rejected before session start.
- Default non-detach `run/` and `run/headless/` leaves remain unchanged.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("detach setup: tty-watch binary not built (root Setup skipped?)")
	}
	return nil
}
```