# Scenario

**Feature**: resume-ready (bound+exited) classifies as resume

```
# session meta exists; probe ResumeReady=true (parity with probeSessionStatus)
Classify(store, id, probe{ResumeReady=true})
  -> ModeResume, found=true
```

## Preconditions

- Seed meta with runner_session_id (bound).
- Probe sets ResumeReady true (exited may also be true).

## Steps

1. Seed finished/bound session.
2. Install resume-ready probe.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-resume-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = "grok-resume-1"
	req.TerminalSessionID = "term-resume-1"
	req.MetaStatus = "finished"
	req.UseProbe = true
	req.ResumeReady = true
	req.RunnerExited = "exited"
	return nil
}
```
