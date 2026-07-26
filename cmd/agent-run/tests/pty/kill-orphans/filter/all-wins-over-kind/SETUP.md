# Scenario

**Feature**: `--all` wins over `--kind` — kind predicates are ignored when `--all` is set

```
spawn child without TestGenerated marker (PPID ≠ 1)
agent-run pty kill-orphans --dry-run --all --kind=test-generated --exe <testbin>
  -> still lists child PID (all wins; kind would have excluded it)
```

## Preconditions

- Child session/path must **not** contain `TestGenerated` so pure kind filter
  would miss it; `--all` must still select it.
- `--exe` isolation.

## Steps

1. Spawn non-orphan serve with plain session id (no TestGenerated).
2. Dry-run with both `--all` and `--kind=test-generated`.
3. Assert child PID is listed; trailing `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "kill-orphans"
	req.SpawnPlan = []ServeSpawnSpec{
		{Label: "child", Orphan: false, SessionID: "pty-all-wins-plain"},
	}
	req.Args = []string{
		"pty", "kill-orphans",
		"--dry-run",
		"--all",
		"--kind=test-generated",
		"--exe", req.AgentRun,
	}
	return nil
}
```
