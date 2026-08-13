# Scenario

**Feature**: Codex ancestor is not a grok hit

```
# codex 2000
#   bash 5000
#     start 6000
#       decoy grok 7000 (session)
FindAncestorGrok(6000) -> ok=false
ResolveFromAncestors -> Kind=none (not Kind=codex)
```

## Preconditions

- Codex is a runner for `ResolveFromPID` but **not** `IsGrokRunner`.
- Descendant decoy grok would fool a descendant walk.

## Steps

1. Set `PID=6000`.
2. Install codex→bash→start plus decoy child grok.

## Context

- Ancestor API is grok-only; do not return `Kind=codex`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	req.Procs = []FixtureProc{
		{PID: pidCodex, PPID: 1, Cmd: "/usr/local/bin/codex"},
		{PID: pidBash, PPID: pidCodex, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
		{PID: pidDecoy, PPID: pidStart, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		pidDecoy: {grokSessionPath(fixtureDecoyGrokSessionID)},
	}
	return nil
}
```
