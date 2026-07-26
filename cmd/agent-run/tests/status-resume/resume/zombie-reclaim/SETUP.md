# Scenario

**Bug**: after `/exit`, zombie keep-alive still holds the TTY registry id (often
equal to agent session id when `--session-id` was used). `resume` must reclaim
that zombie id instead of failing with `session id already in use`.

```
# bound + exited (status resume.ready) but registry still live (zombie)
seed meta session_id == terminal_session_id
  + registry: alive detached serve PID + reachable + exit scrollback
  -> agent-run resume [ --open ] <session-id> ["followup"]
  -> reclaim zombie terminal id
  -> reserve same id (or succeed past reserve)
  -> must NOT: run: session id "…" already in use
```

## Preconditions

- Gate must open: `runner_session_id` bound ∧ `runner.exited` true (zombie
  exit markers / not truly live sendable).
- Registry entry for the **same** id as `SessionID` must be treated as live by
  today's `sessionIDInUse` (alive PID and/or reachable listen_addr) so current
  code RED-fails with "already in use" before reclaim lands.
- Leaves use a **detached** serve PID (`startDetachedSleepPID`) so reclaim may
  kill the process without harming the test runner.
- Fake ptywrap scrollback carries post-exit markers (`grok --resume` +
  `[Terminal exited]`).

## Steps

1. Leaf sets SessionID (often == TerminalSessionID), starts detached serve PID.
2. `seedZombieServeAfterExit` seeds meta + registry + exit scrollback.
3. Run resume (headless followup or `--open` + prompt) with stub TUI / argv probe.
4. Assert: no "already in use"; path proceeds (prefer exit 0).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Grouping: zombie registry reclaim path for resume.
	// Leaves finalize SessionID, args, and runner stubs.
	req.Mode = "read-meta"
	return nil
}
```
