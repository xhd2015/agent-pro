# Scenario

**Feature**: web codex-tty session PTY lifecycle and terminal_session_id mapping

```
POST codex-tty -> HeadlessRun PTY -> meta.terminal_session_id
keep-tty -> terminal available running + finished
follow-up -> same registry entry
```

## Preconditions

- Fake codex TUI via `AGENT_RUN_CODEX_TTY_COMMAND` for deterministic lifecycle.
- Web server started before session creation.

## Steps

1. Grouping setup sets `req.Area = "lifecycle"` and `req.Runner = "codex-tty"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Area = "lifecycle"
	req.Runner = "codex-tty"
	return nil
}
```
