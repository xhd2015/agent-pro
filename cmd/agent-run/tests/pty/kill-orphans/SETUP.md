# Scenario

**Feature**: `agent-run pty kill-orphans` lists or terminates selected agent-run `__serve*` processes

```
# default selection: PPID == 1 only (stricter than pre-filter product)
agent-run pty kill-orphans --dry-run --exe <testbin>
  -> only PPID==1 serves for that exe

# dry-run lists targets without killing
agent-run pty kill-orphans --dry-run --exe <testbin>
  -> print each matching serve (pid, session, command)
  -> process still alive

# empty match set
agent-run pty kill-orphans --dry-run --exe <unique-path>
  -> "no orphans" / "no matching serves"; exit 0

# real kill (tests always pass --exe for isolation)
agent-run pty kill-orphans --exe <testbin>
  -> terminate matching serves; process gone

# filter subtree: --kind / --all / unknown kind (see filter/)
```

## Preconditions

- Live kill/dry-run leaves that spawn serves use Mode `kill-orphans`.
- **Always** pass `--exe` to the test binary path when a single path covers the
  spawned set — never bare `kill-orphans` against host brainstorm/seatalk serves.
- Existing leaves spawn **double-fork orphans (PPID 1)** so they stay valid under
  the stricter default filter.

## Steps

1. Grouping sets Mode `kill-orphans` default (leaves override for help-only).
2. Leaves set Args, SpawnServe / SpawnPlan, and exe filter.
3. `Run` optionally spawns serve(s) then runs CLI (optional follow-up CLI).
4. Assert exit code, stdout, and process liveness / selection.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Default mode for live kill leaves; help-documents-flags overrides Mode "".
	if req.Mode == "" {
		req.Mode = "kill-orphans"
	}
	return nil
}
```
