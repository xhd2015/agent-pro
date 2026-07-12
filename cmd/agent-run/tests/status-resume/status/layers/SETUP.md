# Scenario

**Feature**: multi-layer status probe — session / process / terminal / runner / resume

```
seed meta + optional registry/ptywrap
  -> agent-run status <id> [--json]
  -> runner.exited + resume.ready derived from gate
  # branches: dead terminal | zombie serve (exit markers) | live sendable | unbound
```

## Steps

1. Leaf seeds a specific runner/terminal state (dead, zombie-after-exit, live, unbound).
2. Run status (human or JSON).
3. Assert layer fields and resume gate.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Multi-layer probe leaves seed meta; default runner grok-tty.
	req.Runner = defaultRunner
	return nil
}
```
