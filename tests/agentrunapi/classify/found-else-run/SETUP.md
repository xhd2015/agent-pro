# Scenario

**Feature**: found session that is neither live nor resume-ready → run

```
# meta exists but unbound / exited unknown / not ResumeReady
Classify(store, id, probe{ResumeReady=false, RunnerExited=nil})
  -> ModeRun, found=true
```

## Preconditions

- Seed meta without requiring runner_session_id.
- Probe returns not ready and exited unknown (else branch of CLI classify).

## Steps

1. Seed session meta (unbound).
2. Probe: ResumeReady false, RunnerExited unknown.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-unbound-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = "" // unbound
	req.TerminalSessionID = ""
	req.MetaStatus = "running"
	req.UseProbe = true
	req.ResumeReady = false
	req.RunnerExited = "unknown"
	return nil
}
```
