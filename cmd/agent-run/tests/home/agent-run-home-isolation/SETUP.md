# Scenario

**Feature**: all durable writes stay under `AGENT_RUN_HOME`

```
agent-run run --json → no files outside AGENT_RUN_HOME under temp dir
```

## Preconditions

- `fake-codex` on PATH (inherited from root `SETUP.md`).
- `AGENT_RUN_HOME` set to isolated path under `t.TempDir()`.

## Steps

1. Run `agent-run run --json --agent-runner fake-codex "hi"` with isolated `AGENT_RUN_HOME`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"run", "--json", "--agent-runner", "fake-codex", "hi"}
	return nil
}
```