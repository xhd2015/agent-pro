# Scenario

**Feature**: `--codex` conflicts with explicit `--agent-runner=grok-tty`

```
# --codex aliases --agent-runner=codex-tty; grok-tty is a mismatch
agent-run takeover --codex --agent-runner grok-tty <provider-session-id>
  -> exit non-zero
  -> conflict / mismatch / mutually exclusive
```

## Steps

1. Pass `--codex` with mismatched long-form runner and a session id.
2. Run Mode handle.
3. Assert non-zero exit and conflict wording (not unknown command).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{
		"takeover",
		"--codex",
		"--agent-runner", "grok-tty",
		takeoverFixtureSessionID,
	}
	return nil
}
```
