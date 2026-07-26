# Scenario

**Feature**: found session, Probe nil → RunSession once (LifecycleProbe default)

```
AutoSendOrResume(seeded session, Probe=nil, hooks)
  -> Classify via LifecycleProbe (unknown on temp store)
  -> ModeRun
  -> RunSession x1; SendLive=0; Resume=0
# no agent-run LookPath
```

## Preconditions

- Session seeded (found=true); Probe left nil (not EmptyProbe, not script).
- InstallHooks from parent dispatch SETUP.
- No real TTY registry → not ModeSend.

## Steps

1. Seed unbound meta.
2. ProbeName=nil; UseProbe false.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "sess-auto-nil-probe"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = ""
	req.TerminalSessionID = ""
	req.MetaStatus = "running"
	req.UseProbe = false
	req.ProbeName = "nil"
	req.Prompt = "follow up after nil probe"
	return nil
}
```
