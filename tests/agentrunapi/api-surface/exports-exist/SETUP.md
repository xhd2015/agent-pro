# Scenario

**Feature**: Mode constants and entrypoints exist and are callable

```
agentrunapi.ModeRun/Send/Resume + Classify + AutoSendOrResume + Opts + ProbeReport
  -> compile + callable (classify missing id; auto empty session validates)
```

## Preconditions

- Empty store home is enough; no seed, no probe, no hooks.

## Steps

1. Rely on root Home and mode from grouping.
2. Assert symbols usable and Classify returns ModeRun for missing session.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Mode already api_surface; no seed.
	req.SeedMeta = false
	req.SessionID = "api-surface-missing"
	return nil
}
```
