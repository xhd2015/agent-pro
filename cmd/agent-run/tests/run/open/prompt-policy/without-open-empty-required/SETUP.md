# Scenario

**Feature**: without `--open`, empty prompt still fails (compat regression)

```
agent-run run --agent-runner fake-codex
  -> exit ≠ 0
  -> "prompt is required"
```

## Steps

1. Run with a runner and no prompt args (and no `--open`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "fake-codex"
	// No prompt positional; no --open.
	req.Args = []string{"run", "--agent-runner", "fake-codex"}
	return nil
}
```
