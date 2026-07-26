# Scenario

**Feature**: live session dispatches SendLive only

```
AutoSendOrResume(live probe, hooks)
  -> ModeSend
  -> SendLive x1; Run=0; Resume=0
```

## Preconditions

- Seeded meta + live probe (`RunnerExited=false`).
- Hooks installed.

## Steps

1. Seed live session; inject live probe.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-auto-send-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = "grok-auto-send"
	req.TerminalSessionID = "term-auto-send"
	req.MetaStatus = "running"
	req.UseProbe = true
	req.ResumeReady = false
	req.RunnerExited = "live"
	req.Prompt = "follow up please"
	return nil
}
```
