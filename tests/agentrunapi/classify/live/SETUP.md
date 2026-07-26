# Scenario

**Feature**: live runner (exited=false) classifies as send

```
# session meta exists; probe says runner still active
Classify(store, id, probe{RunnerExited=false})
  -> ModeSend, found=true
```

## Preconditions

- Seed meta for session (runner bound optional; live uses Exited=false).
- Probe injects `RunnerExited=false` and `ResumeReady=false`.

## Steps

1. Seed session meta.
2. Install live probe.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-live-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = "grok-live-1"
	req.TerminalSessionID = "term-live-1"
	req.MetaStatus = "running"
	req.UseProbe = true
	req.ResumeReady = false
	req.RunnerExited = "live"
	return nil
}
```
