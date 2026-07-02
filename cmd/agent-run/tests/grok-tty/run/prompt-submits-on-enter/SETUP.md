# Scenario

**Bug**: prompt injected with bare `\n` appears in grok input but is never submitted
(like typing without pressing Enter). Interactive TUIs expect carriage return `\r`.

```
fake TUI submits only on \r → run "run ls" → must capture SUBMITTED:run ls
```

Reproduces user report: `agent-run run --agent-runner=grok-tty "run ls"` then attach
shows `run ls` in the input buffer without grok executing.

## Preconditions

- Fake TUI uses `fakeTUIRequiresCR()` — echoes chars, submits on `\r` only; bare `\n`
  prints `UNSUBMITTED:<line>` and exits.

## Steps

1. Set `AGENT_RUN_GROK_TTY_COMMAND` to CR-only fake TUI.
2. Run `agent-run run --agent-runner grok-tty "run ls"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	setGrokTTYCommand(req, fakeTUIRequiresCR())
	req.Args = append(req.Args, "run ls")
	return nil
}
```