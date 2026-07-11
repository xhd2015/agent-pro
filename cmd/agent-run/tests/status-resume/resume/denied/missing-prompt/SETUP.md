# Scenario

**Feature**: resume without followup and without `--open` requires a prompt

```
exited bound meta -> agent-run resume <id>  (no followup, no --open)
  -> exit 1, prompt required
```

## Steps

1. Seed bound exited meta.
2. Run resume with only session id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
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
