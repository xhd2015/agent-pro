# Scenario

**Feature**: resume accepts `--open` (not unknown-flag); gate may still apply

```
seed bound+exited meta
  -> agent-run resume --open <id>
  -> must not fail with "unknown flag" / "unknown option" for --open
  -> empty prompt allowed under --open (mirrors run policy)
```

## Steps

1. Seed exited bound meta.
2. Run `resume --open <id>` with instant-attach + quick fake TUI so the path
   can proceed if gates pass (no followup required with `--open`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-open-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440600"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-open-1"
	req.InitialPrompt = "prior open"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	req.OpenInstantAttach = true
	req.GrokTTYCommand = fakeTUIRespondHi()
	req.Args = []string{"resume", "--open", req.SessionID}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
