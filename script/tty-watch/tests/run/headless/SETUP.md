# Scenario

**Feature**: `tty-watch run --headless` starts a PTY session without writer attach

```
# headless parent: reserve id, spawn serve child, skip attachWriter
harness pipe stdin -> tty-watch run --headless <cmd> -> registry session-N.json
tty-watch -> stdout: session-id: <id>\n -> block until child exit or SIGINT
# SIGINT: forward to PTY child, wait <=10s, stderr status after 1s, force-kill exit 1
tty-watch list -> session live while headless parent blocked
```

## Preconditions

- Isolated `TTY_WATCH_HOME` from root Setup.
- Headless runs use **pipe stdin** (not PTY attach); harness reads the first stdout line for session id.
- `buildHeadlessRunArgv` always inserts `--headless` before optional `--session-id` and `<command>`.

## Steps

1. Leaf sets `req.Phase` to a `run-headless-*` key and optional `RunCommand` / `CustomSessionID`.
2. Grouping helpers start headless via pipes, parse `session-id: <id>\n`, and optionally SIGINT the headless parent.
3. `Run` (root `DOCTEST.md`) dispatches `run-headless-*` phases to helpers below before `ttywatchtest.Run`.

## Context

- Natural child exit: headless parent exits 0; registry cleaned by serve-child grace logic.
- Ctrl-C on headless parent: forward INT to inner program; single stderr line after 1s if still running.
- Default non-headless `run/` leaves remain unchanged (writer attach, silent host).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("headless setup: tty-watch binary not built (root Setup skipped?)")
	}
	return nil
}
```