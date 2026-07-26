# Scenario

**Feature**: explicit `--session` on grok-tty uses the same id for registry and storage

```
agent-run run --agent-runner grok-tty --keep-tty --session my-task "hi"
  -> stderr grok-tty: my-task
  -> sessions/grok-tty/my-task/
  -> meta.terminal_session_id == my-task
  -> grok-tty-registry/my-task.json
```

## Steps

1. Configure fake TUI respond-hi.
2. Run with `--session my-task` and prompt `hi`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	req.KeepTTY = true
	req.Prompt = "hi"
	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = append(req.Args,
		"--agent-runner", "grok-tty",
		"--keep-tty",
		"--session", "my-task",
		req.Prompt,
	)
	return nil
}
```
