# Scenario

**Feature**: `--no-submit` without `--open` is rejected

```
agent-run run --agent-runner grok-tty --no-submit "x"
  -> exit ≠ 0
  -> error explains --no-submit requires --open
```

## Preconditions

- TTY runner + prompt used so the failure is flag pairing, not missing prompt /
  non-TTY rejection alone.
- Fake TUI optional — validation should fail before long PTY work.

## Steps

1. Invoke run with `--no-submit`, a TTY runner, and a short prompt (no `--open`).
2. Assert non-zero exit and clear requires-`--open` wording.

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
		"--no-submit",
		req.Prompt,
	}
	return nil
}
```
