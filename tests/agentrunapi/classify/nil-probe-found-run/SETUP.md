# Scenario

**Feature**: found session + nil probe → LifecycleProbe default → ModeRun

```
# meta exists; probe=nil → Classify uses LifecycleProbe (not EmptyProbe)
# temp store without TTY registry → unknown/not live → ModeRun
Classify(store, id, nil)
  -> ModeRun, found=true (no crash)
```

## Preconditions

- Seed meta; do **not** inject a script or EmptyProbe.
- Unit-testable contract: empty/temp store + seeded meta + no TTY registry
  still yields ModeRun (not ModeSend).

## Steps

1. Seed unbound session under temp home.
2. ProbeName empty / UseProbe false → nil probe.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "sess-nil-probe-1"
	req.SeedMeta = true
	req.Runner = "grok-tty"
	req.RunnerSessionID = ""
	req.TerminalSessionID = ""
	req.MetaStatus = "running"
	req.UseProbe = false
	req.ProbeName = "nil"
	return nil
}
```
