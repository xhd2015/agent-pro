# Scenario

**Feature**: `status --grok-session-id` finds a finished `grok-tty` session by
provider UUID (CLI `--agent-runner` is ignored for lookup)

```
seed sessions/<id>/meta.json runner=grok-tty runner_session_id=UUID finished bound
  -> agent-run status --agent-runner fake-codex --grok-session-id UUID
  -> exit 0; stdout identifies grok-tty/<session_id>
```

## Steps

1. Seed finished bound+exited meta for `grok-tty` with a known UUID.
2. Run status with `--grok-session-id` only (no positional). Pass a conflicting
   `--agent-runner` to prove meta-only filter.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	req.SessionID = "test-gsid-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440901"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-gsid-1"
	req.InitialPrompt = "grok-session-id resolve"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	// --agent-runner is ignored for lookup (meta.runner still grok-tty).
	req.Args = []string{
		"status",
		"--agent-runner", "fake-codex",
		"--grok-session-id", req.RunnerSessionID,
	}
	return nil
}
```
