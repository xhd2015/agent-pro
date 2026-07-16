# Scenario

**Feature**: `resume --detach --json` is rejected

```
seed bound+exited
  -> agent-run resume --detach --json <id> "hi"
  -> exit ≠ 0; exclusivity error
```

## Steps

1. Seed exited bound meta.
2. Invoke resume with both flags and a short followup.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-detach-json"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440702"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-detach-json"
	req.InitialPrompt = "prior"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = []string{
		"resume",
		"--detach",
		"--json",
		req.SessionID,
		"hi",
	}
	return nil
}
```
