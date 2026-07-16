# Scenario

**Feature**: `--agent-runner codex-tty` is rejected for `--resume-from-grok-session`
(only `grok-tty` allowed when the runner flag is set)

```
seed Grok session UUID under GROK_HOME
  -> agent-run run --agent-runner codex-tty --resume-from-grok-session UUID
  -> exit 1; requires grok-tty
```

## Preconditions

- Valid Grok fixture seeded so failure is the runner gate (not missing session),
  once validation order reaches runner/lookup.

## Steps

1. Seed Grok session at process workspace (`req.WorkDir`).
2. Run with `--agent-runner codex-tty`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	req.AgentRunner = "codex-tty"
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
