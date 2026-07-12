# Scenario

**Feature**: `--all` includes non-PPID1 serves; default does not

```
spawn child (PPID ≠ 1): testbin __serve_*__ …
agent-run pty kill-orphans --dry-run --all --exe <testbin>
  -> lists child PID
follow-up: agent-run pty kill-orphans --dry-run --exe <testbin>
  -> does not list child
```

## Preconditions

- Non-orphan child under test binary; `--exe` isolation.
- Follow-up contrasts default PPID1 filter.

## Steps

1. Spawn non-orphan serve.
2. Dry-run with `--all --exe`.
3. Follow-up default dry-run with same `--exe`.
4. Assert `--all` lists PID; default does not.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnPlan = []ServeSpawnSpec{
		{Label: "child", Orphan: false, SessionID: "pty-all-child"},
	}
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--all",
		"--exe", req.AgentRun,
	}
	req.FollowUpArgs = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--exe", req.AgentRun,
	}
	return nil
}
```
