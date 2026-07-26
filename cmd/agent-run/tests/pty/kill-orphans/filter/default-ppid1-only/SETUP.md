# Scenario

**Feature**: default kill-orphans (no `--kind`, no `--all`) lists only PPID==1 serves

```
spawn orphan (double-fork, PPID 1):  testbin __serve_*__ sess-ppid1 sleep N
spawn child  (harness child, PPID≠1): testbin __serve_*__ sess-child sleep N
agent-run pty kill-orphans --dry-run --exe <testbin>
  -> stdout lists orphan PID
  -> stdout does not list child PID
  -> both processes still alive
```

## Preconditions

- Same `--exe` path for both serves so isolation is tight.
- Mode `kill-orphans` with dual `SpawnPlan`.

## Steps

1. Spawn one PPID1 orphan and one non-orphan child of the harness.
2. Run default dry-run with `--exe` = test binary.
3. Assert only the orphan PID is listed; child absent; trailing `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnServe = false
	req.SpawnPlan = []ServeSpawnSpec{
		{Label: "ppid1", Orphan: true, SessionID: "pty-filter-ppid1"},
		{Label: "child", Orphan: false, SessionID: "pty-filter-child"},
	}
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--exe", req.AgentRun,
	}
	return nil
}
```
