# Scenario

**Feature**: found session + EmptyProbe → ModeRun (unknown lifecycle)

```
# meta exists; EmptyProbe returns ResumeReady=false, RunnerExited=nil
Classify(store, id, EmptyProbe)
  -> ModeRun, found=true
```

## Preconditions

- Seed meta under temp FileStore (no TTY registry required).
- Explicit `EmptyProbe` (not nil default, not script probe).

## Steps

1. Seed unbound session meta.
2. ProbeName=empty → package EmptyProbe.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-empty-probe-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = ""
	req.TerminalSessionID = ""
	req.MetaStatus = "running"
	req.UseProbe = false
	req.ProbeName = "empty"
	return nil
}
```
