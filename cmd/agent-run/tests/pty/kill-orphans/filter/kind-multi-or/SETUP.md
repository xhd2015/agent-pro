# Scenario

**Feature**: multiple `--kind` values are OR-ed (no PPID filter)

```
spawn child A: session id contains TestGenerated (same temp exe path)
spawn child B: plain session; path under /var/folders/…/T/ (workdir-at-tmp)
agent-run pty kill-orphans --dry-run \
  --kind=test-generated --kind=workdir-at-tmp --exe <testbin>
  -> stdout lists both A and B PIDs
```

## Preconditions

- Skip when temp layout is not workdir-at-tmp (B must match that kind).
- Both children PPID ≠ 1; shared `--exe` for isolation.
- A matches `test-generated` via session/argv `TestGenerated`; B matches
  workdir-at-tmp via binary path only (session without TestGenerated).

## Steps

1. Skip if TempDir is not macOS workdir-at-tmp layout.
2. Spawn two non-orphan serves under the same test binary.
3. Dry-run with both kinds and `--exe`.
4. Assert both PIDs appear in stdout; trailing `\n`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if !tempDirLooksLikeWorkdirAtTmp(req.TempDir) {
		t.Skipf("kind-multi-or needs workdir-at-tmp temp layout; TempDir=%s", req.TempDir)
	}
	req.Mode = "kill-orphans"
	req.SpawnPlan = []ServeSpawnSpec{
		{
			Label:     "tg",
			Orphan:    false,
			SessionID: "pty-multi-TestGenerated-a",
		},
		{
			Label:     "tmp",
			Orphan:    false,
			SessionID: "pty-multi-workdir-b",
		},
	}
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--kind=test-generated",
		"--kind=workdir-at-tmp",
		"--exe", req.AgentRun,
	}
	return nil
}
```
