# Scenario

**Feature**: `agent-run pty kill-orphans` lists or terminates agent-run `__serve*` processes

```
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
```

## Preconditions

- Live kill/dry-run leaves that spawn serves use Mode `kill-orphans`.
- **Always** pass `--exe` to the test binary path — never bare `kill-orphans`
  in this tree (would hit host brainstorm/seatalk serves).

## Steps

1. Grouping sets Mode `kill-orphans` default (leaves override for help-only).
2. Leaves set Args, SpawnServe, and exe filter.
3. `Run` optionally spawns serve then runs CLI.
4. Assert exit code, stdout, and process liveness.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Default mode for live kill leaves; help-documents-flags overrides Mode "".
	if req.Mode == "" {
		req.Mode = "kill-orphans"
	}
	return nil
}
```
