# Scenario

**Feature**: `kill-orphans --exe` terminates only serves of the test binary

```
spawn: testbin __serve_sleep_N__ <session> sleep N
agent-run pty kill-orphans --exe <testbin>
  -> process no longer alive
  -> exit 0; summary on stdout with trailing \n
```

## Preconditions

- Uses `--exe` so unrelated host serves are not killed.
- Mode `kill-orphans` with `SpawnServe=true`.

## Steps

1. Spawn detached serve re-exec of the test binary.
2. Run `pty kill-orphans --exe <testbin>` (no dry-run).
3. Assert exit 0, serve dead, trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnServe = true
	req.ServeHoldSecs = 120
	req.Args = []string{
		"pty", "kill-orphans",
		"--exe", req.AgentRun,
	}
	return nil
}
```
