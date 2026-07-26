# Scenario

**Feature**: AutoSendOrResume dispatches to mode-specific hooks (no agent-run binary)

```
Classify -> Mode
  ModeRun    -> RunSession once
  ModeSend   -> SendLive once
  ModeResume -> ResumeSession once
# NewTerminal=false; hooks cover path → no LookPath("agent-run")
```

## Preconditions

- InstallHooks true on every dispatch leaf.
- Probe scripts mirror classify leaves.
- Production send/run/resume must not run when hooks set.
- Harness forces `Opts.NewTerminal=false` (library in-process path).

## Steps

1. Enable hooks; leaves seed meta + probe.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InstallHooks = true
	req.ExpectNoHooks = false
	return nil
}
```
