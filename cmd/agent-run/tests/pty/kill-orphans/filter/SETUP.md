# Scenario

**Feature**: kill-orphans selection filters — default PPID1, `--kind`, `--all`

```
# default: only PPID == 1
agent-run pty kill-orphans --dry-run --exe <bin>
  -> list PPID1 serves; exclude harness children (PPID ≠ 1)

# kind replaces PPID filter (OR when multiple)
agent-run pty kill-orphans --dry-run --kind=test-generated --exe <bin>
agent-run pty kill-orphans --dry-run --kind=workdir-at-tmp --exe <bin>
agent-run pty kill-orphans --dry-run --kind=test-generated --kind=workdir-at-tmp --exe <bin>

# --all wins over --kind; lists every agent-run __serve* (still --exe ANDed)
agent-run pty kill-orphans --dry-run --all --exe <bin>
agent-run pty kill-orphans --dry-run --all --kind=test-generated --exe <bin>

# unknown kind rejected
agent-run pty kill-orphans --kind=not-a-real-kind -> exit 1
```

## Preconditions

- Mode `kill-orphans` (inherited from parent grouping when empty).
- Live leaves use `--exe` isolation against the leaf test binary path.
- `workdir-at-tmp` leaves skip when `req.TempDir` is not under `/var/folders/…/T/`.
- Prefer `SpawnPlan` with `Orphan: false` for non-PPID1 children; `Orphan: true`
  for true orphans (double-fork).

## Steps

1. Filter leaves set `SpawnPlan` / `Args` / optional `FollowUpArgs`.
2. `Run` spawns serves, runs primary CLI, optional follow-up default dry-run.
3. Assert which PIDs appear in stdout and trailing newline / exit codes.

## Context

- Selection order: all → kind OR → else PPID1; then `--exe`; drop self.
- Kind markers: `TestGenerated` in path or argv; `/var/folders/` + `/T/` for workdir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Mode == "" {
		req.Mode = "kill-orphans"
	}
	if req.ServeHoldSecs <= 0 {
		req.ServeHoldSecs = 120
	}
	return nil
}
```
