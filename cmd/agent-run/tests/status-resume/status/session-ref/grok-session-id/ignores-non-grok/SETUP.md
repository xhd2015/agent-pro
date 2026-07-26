# Scenario

**Feature**: non-grok runners never match `--grok-session-id` even when UUID equals
their `runner_session_id`

```
seed runner=codex-tty runner_session_id=UUID
  -> agent-run status --grok-session-id UUID
  -> exit 1; not found (must not resolve codex)
```

## Steps

1. Seed only a `codex-tty` session with the target UUID.
2. Run status with `--grok-session-id` for that UUID.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "test-gsid-codex-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440905"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-gsid-codex-1"
	req.InitialPrompt = "codex must not match"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	req.Args = []string{"status", "--grok-session-id", req.RunnerSessionID}
	return nil
}
```
