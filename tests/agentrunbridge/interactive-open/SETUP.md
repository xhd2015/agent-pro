# Scenario

**Feature**: `RunInteractiveOpen` fills assumed flags and calls `Run`

```
InteractiveOpenOpts
  -> fill AutoSendOrResume+NewTerminal+Open+WaitReady; AgentRunner default grok-tty
  -> Run(RunOpts) only (single exec path)
  -> wait-ready via tty status polls
```

## Preconditions

- Required: non-empty SessionID + Prompt.
- Optional: WorkspaceDir, NoSubmit, Binary, AgentRunner, hooks, ReadyTimeout/Poll.
- Default AgentRunner when empty: `grok-tty`.
- Leaves inject fakes; status polls scripted ready unless noted.

## Steps

1. Set Mode to `interactive_open` (composition leaf overrides to `interactive_open_vs_run`).
2. Leaf sets minimal or extended InteractiveOpen fields.
3. Assert launch argv, runner default, wait success, or argv parity with BuildArgs fill.

## Context

- Assumed mapping documented in root DOCTEST.md DSN.
- Composition leaf uses Mode `interactive_open_vs_run`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Default mode for this branch; composition leaf overwrites.
	if req.Mode == "" {
		req.Mode = "interactive_open"
	}
	// Wait-ready always true for InteractiveOpen — script ready by default so
	// leaves that only care about argv finish without timeout.
	if len(req.StatusPollSeq) == 0 && req.StatusPollHold == "" {
		req.StatusPollSeq = []string{statusReadyFixture()}
	}
	return nil
}
```
