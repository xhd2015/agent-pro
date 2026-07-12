# Scenario

**Feature**: dry-run lists a spawned test `__serve` when filtered by `--exe`

```
spawn: testbin __serve_sleep_N__ <session> sleep N
agent-run pty kill-orphans --dry-run --exe <testbin>
  -> stdout includes spawned PID (or session id)
  -> serve process still alive after dry-run
```

## Preconditions

- Spawn uses the same leaf `req.AgentRun` path as `--exe` so only the test
  binary's serve is listed.
- Mode `kill-orphans` with `SpawnServe=true`.

## Steps

1. Spawn detached serve re-exec of the test agent-run binary.
2. Run dry-run kill-orphans with `--exe` = test binary.
3. Assert exit 0, output mentions PID/session, process still alive, trailing `\n`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnServe = true
	req.ServeHoldSecs = 120
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--exe", req.AgentRun,
	}
	return nil
}
```
