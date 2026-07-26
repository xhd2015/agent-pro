# Scenario

**Feature**: `--open` + `--json` on a TTY runner is rejected

```
agent-run run --agent-runner grok-tty --open --json "x"
  -> exit ≠ 0
  -> stderr explains mutual exclusion / incompatibility
```

## Steps

1. Invoke run with both flags on `grok-tty` and a short prompt.
2. Fake TUI is optional — validation should fail before long PTY work.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	req.Prompt = "x"
	// Install fake TUI in case validation order starts the runner first.
	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"--open",
		"--json",
		req.Prompt,
	}
	return nil
}
```
