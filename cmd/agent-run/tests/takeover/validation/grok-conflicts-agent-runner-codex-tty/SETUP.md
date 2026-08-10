# Scenario

**Feature**: `--grok` conflicts with explicit `--agent-runner=codex-tty`

```
# --grok aliases --agent-runner=grok-tty; codex-tty is a mismatch
agent-run takeover --grok --agent-runner codex-tty <provider-session-id>
  -> exit non-zero
  -> conflict / mismatch / mutually exclusive
```

## Steps

1. Pass `--grok` with mismatched long-form runner and a session id.
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
		"--grok",
		"--agent-runner", "codex-tty",
		takeoverFixtureSessionID,
	}
	return nil
}
```
