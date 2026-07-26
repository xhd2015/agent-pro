# Scenario

**Feature**: `--grok-session-id` and positional `<session-id>` are mutually exclusive

```
seed session
  -> agent-run status --grok-session-id UUID <session-id>
  -> exit 1; mutually exclusive error
```

## Steps

1. Seed a valid finished bound session (so failure is mutex, not missing meta).
2. Pass both flag and positional session id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	req.SessionID = "test-gsid-mutex-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440906"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-gsid-mutex-1"
	req.InitialPrompt = "mutex"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	req.Args = []string{
		"status",
		"--grok-session-id", req.RunnerSessionID,
		req.SessionID,
	}
	return nil
}
```
