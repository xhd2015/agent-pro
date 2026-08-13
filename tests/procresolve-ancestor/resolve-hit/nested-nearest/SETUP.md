# Scenario

**Feature**: nested groks — nearest (subagent) wins, not the topmost main grok

```
# main grok 3000  Lsof -> main uuid
#   subagent grok 4242  Lsof -> nearest uuid
#     bash 5000
#       start 6000
FindAncestorGrok(6000) -> 4242
ResolveFromAncestors(6000) -> SessionID=nearest, not main
```

## Preconditions

- Both groks have parseable session paths.
- Nearest is the first `IsGrokRunner` walking start then PPID.

## Steps

1. Set `PID=6000`.
2. Install main→sub→bash→start.
3. OpenFiles for both groks.

## Context

- Topmost main uuid must not win.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	req.Procs = []FixtureProc{
		{PID: pidMainGrok, PPID: 1, Cmd: "/usr/local/bin/grok --resume " + fixtureMainGrokSessionID},
		{PID: pidGrok, PPID: pidMainGrok, Cmd: "/usr/local/bin/grok"},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
	}
	req.OpenFiles = map[int][]string{
		pidMainGrok: {grokSessionPath(fixtureMainGrokSessionID)},
		pidGrok:     {grokSessionPath(fixtureGrokSessionID)},
	}
	return nil
}
```
