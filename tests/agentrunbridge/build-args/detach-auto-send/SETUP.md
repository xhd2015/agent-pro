# Scenario

**Feature**: detach-profile RunOpts → argv with auto-send and --detach (no open/new-terminal)

```
BuildArgs(session + grok-tty + auto-send + Detach + prompt)
  -> run --session-id=… --agent-runner=grok-tty --auto-send-or-resume --detach -- <prompt>
  # no --open, no --new-terminal
```

## Preconditions

- Detach-profile flags set on `RunOpts` (as `RunInteractiveDetach` would fill).
- No `WorkspaceDir`, no `NoSubmit`, no `AllowRelocateResumeSessionDir`.
- Classic RED until `RunOpts.Detach` and `BuildArgs` emit `--detach` with
  open-style `--` prompt placement.

## Steps

1. Set session, runner `grok-tty`, auto-send, Detach, prompt.
2. Leave Open and NewTerminal false.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-detach-1"
	req.Prompt = "detach me"
	req.AgentRunner = "grok-tty"
	req.AutoSendOrResume = true
	req.NewTerminal = false
	req.Open = false
	req.Detach = true
	return nil
}
```
