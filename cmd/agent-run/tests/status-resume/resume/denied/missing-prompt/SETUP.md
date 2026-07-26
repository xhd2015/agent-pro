# Scenario

**Feature**: resume without followup is allowed (reopen only; not an error)

```
exited bound meta -> agent-run resume <id>  (no followup, no --open)
  -> does NOT fail with "prompt is required"
  -> keep-tty reopen path (may still error on missing stub binary in this leaf)
```

## Steps

1. Seed bound exited meta.
2. Run resume with only session id (no followup prompt).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-resume-noprompt-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440422"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-np-1"
	req.InitialPrompt = "prior"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	req.Args = []string{"resume", req.SessionID}
	return nil
}
```
