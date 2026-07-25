# Scenario

**Feature**: hard session resolve when a grok or codex runner has open session files

```
# candidates found; Lsof yields parseable session path
ResolveFromPID -> Kind=grok|codex, Confidence=hard, SessionID=<uuid>
```

## Preconditions

- At least one grok/codex candidate exists in the tree (not `grok update`).
- Fake Lsof returns a path containing a session uuid for the winning runner.

## Steps

1. Leaves install multi- or single-node fixtures and open-file paths.
2. Assert Kind, SessionID, Confidence=hard, Source, RunnerPID.

## Context

- Source is `open-files` when the input pid itself is the runner.
- Source is `open-files+tree` when the runner is a descendant of the input pid.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Tag branch for debugging harness dumps; leaves override Procs/OpenFiles/PID.
	if req.GrokHome == "" {
		req.GrokHome = "/tmp/fake-grok-home"
	}
	return nil
}
```
