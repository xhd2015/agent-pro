# Scenario

**Feature**: found session + EmptyProbe → RunSession once

```
AutoSendOrResume(seeded session, Probe=EmptyProbe, hooks)
  -> Classify ModeRun
  -> RunSession x1; Send=0; Resume=0
```

## Preconditions

- Session seeded; explicit EmptyProbe.
- InstallHooks from parent; no agent-run binary.

## Steps

1. Seed unbound meta.
2. ProbeName=empty.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-auto-empty-probe"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = ""
	req.TerminalSessionID = ""
	req.MetaStatus = "running"
	req.UseProbe = false
	req.ProbeName = "empty"
	req.Prompt = "follow up empty probe"
	return nil
}
```
