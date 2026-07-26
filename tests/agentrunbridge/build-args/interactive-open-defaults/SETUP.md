# Scenario

**Feature**: interactive-open profile RunOpts → default argv shape

```
BuildArgs(session + grok-tty + auto-send + new-terminal + open + prompt)
  -> run --session-id=… --agent-runner=grok-tty --auto-send-or-resume --new-terminal --open -- <prompt>
```

## Preconditions

- Open-profile flags set on `RunOpts` (as `RunInteractiveOpen` would fill).
- No `WorkspaceDir`, no `NoSubmit`.

## Steps

1. Set session, runner `grok-tty`, auto-send, new-terminal, open, prompt.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-open-1"
	req.Prompt = "open me"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.NewTerminal = true
	req.Open = true
	return nil
}
```
