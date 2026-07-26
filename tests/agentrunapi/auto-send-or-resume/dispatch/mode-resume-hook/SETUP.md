# Scenario

**Feature**: resume-ready session dispatches ResumeSession only

```
AutoSendOrResume(ResumeReady probe, hooks)
  -> ModeResume
  -> ResumeSession x1; Run=0; Send=0
```

## Preconditions

- Seeded bound+exited meta; probe ResumeReady true.
- Hooks installed.

## Steps

1. Seed finished/bound session; inject resume-ready probe.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-auto-resume-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = "grok-auto-resume"
	req.TerminalSessionID = "term-auto-resume"
	req.MetaStatus = "finished"
	req.UseProbe = true
	req.ResumeReady = true
	req.RunnerExited = "exited"
	req.Prompt = "resume followup"
	return nil
}
```
