# Scenario

**Feature**: pure CollectTTYsFromTree from process snapshot

```
[]ProcRow + rootPID -> CollectTTYsFromTree -> real TTYs only
```

## Preconditions

- No Store / iTerm / Focus involved.

## Steps

1. Leaves set `Phase=collect-ttys`, `RootPID`, and `Procs`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "collect-ttys"
	return nil
}
```
