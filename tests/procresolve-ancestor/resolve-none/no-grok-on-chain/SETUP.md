# Scenario

**Feature**: no grok on the PPID chain — none, even if a child grok exists

```
# init → bash 5000 → start 6000
#                    └ decoy grok 7000 (Lsof session)
FindAncestorGrok(6000) -> ok=false
ResolveFromAncestors(6000) -> Kind=none
```

## Preconditions

- Descendant decoy is required so `ResolveFromPID(6000)` would hard-hit.

## Steps

1. Set `PID=6000`.
2. Install bash chain + decoy child grok with session path.

## Context

- Ancestor APIs must not walk descendants.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	req.Procs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "/sbin/init"},
		{PID: pidBash, PPID: 1, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
		{PID: pidDecoy, PPID: pidStart, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		pidDecoy: {grokSessionPath(fixtureDecoyGrokSessionID)},
	}
	return nil
}
```
