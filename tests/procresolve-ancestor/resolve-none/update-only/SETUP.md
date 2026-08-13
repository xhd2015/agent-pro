# Scenario

**Feature**: only `grok update` above start is not a grok ancestor

```
# grok update 4300
#   bash 5000
#     start 6000
#       decoy grok 7000
FindAncestorGrok(6000) -> ok=false
ResolveFromAncestors -> Kind=none
```

## Preconditions

- Update utility is skipped; no real grok further up.
- Descendant decoy would fool `ResolveFromPID`.

## Steps

1. Set `PID=6000`.
2. Install update→bash→start plus decoy child.

## Context

- Same classification as `IsGrokRunner` / existing exclude-grok-update.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	req.Procs = []FixtureProc{
		{PID: pidUpdate, PPID: 1, Cmd: "/usr/local/bin/grok update"},
		{PID: pidBash, PPID: pidUpdate, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
		{PID: pidDecoy, PPID: pidStart, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		pidDecoy: {grokSessionPath(fixtureDecoyGrokSessionID)},
	}
	return nil
}
```
