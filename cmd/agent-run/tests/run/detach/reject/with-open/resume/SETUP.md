# Scenario

**Feature**: `resume --detach --open` is rejected

```
seed bound+exited
  -> agent-run resume --detach --open <id>
  -> exit ≠ 0; mutual exclusion
```

## Steps

1. Seed exited bound meta.
2. Invoke resume with both flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-detach-open"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440701"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-detach-open"
	req.InitialPrompt = "prior"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = []string{
		"resume",
		"--detach",
		"--open",
		req.SessionID,
	}
	return nil
}
```
