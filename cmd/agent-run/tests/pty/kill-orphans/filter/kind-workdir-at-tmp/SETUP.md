# Scenario

**Feature**: `--kind=workdir-at-tmp` matches serves under macOS Go temp `/var/folders/…/T/`

```
spawn child: <t.TempDir()>/bin/agent-run __serve_*__ … (PPID ≠ 1)
  # t.TempDir on macOS is under /var/folders/…/T/
agent-run pty kill-orphans --dry-run --kind=workdir-at-tmp --exe <testbin>
  -> stdout lists child PID
follow-up default dry-run --exe <testbin>
  -> does not list child
```

## Preconditions

- Skip when `req.TempDir` does not contain both `/var/folders/` and `/T/`
  (Linux `/tmp` is out of scope for this kind).
- Non-orphan child so default PPID1 filter excludes it.
- Default leaf binary path already sits under `t.TempDir()`.

## Steps

1. Skip if temp layout is not workdir-at-tmp.
2. Spawn non-orphan serve with default test binary.
3. Dry-run with `--kind=workdir-at-tmp`; follow-up default dry-run.
4. Assert kind lists PID; default does not.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if !tempDirLooksLikeWorkdirAtTmp(req.TempDir) {
		t.Skipf("workdir-at-tmp requires /var/folders/…/T/ temp layout; TempDir=%s", req.TempDir)
	}
	req.Mode = "kill-orphans"
	req.SpawnPlan = []ServeSpawnSpec{
		{
			Label:      "tmp",
			Orphan:     false,
			PathMarker: "workdir-at-tmp",
			SessionID:  "pty-kind-workdir",
		},
	}
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--kind=workdir-at-tmp",
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
