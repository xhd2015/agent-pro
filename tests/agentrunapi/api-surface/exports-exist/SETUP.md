# Scenario

**Feature**: Mode constants, entrypoints, and lifecycle probes exist and are callable

```
agentrunapi.ModeRun/Send/Resume + Classify + AutoSendOrResume + Opts + ProbeReport
  + LifecycleProbe + EmptyProbe
  -> compile + callable (classify missing id; probes on empty store; auto empty session validates)
```

## Preconditions

- Empty store home is enough; no seed, no hooks.
- `LifecycleProbe` and `EmptyProbe` are package-level `ProbeFunc` values.

## Steps

1. Rely on root Home and mode from grouping.
2. Assert symbols usable, Classify ModeRun for missing session, probes callable.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Mode already api_surface; no seed.
	req.SeedMeta = false
	req.SessionID = "api-surface-missing"
	return nil
}
```
