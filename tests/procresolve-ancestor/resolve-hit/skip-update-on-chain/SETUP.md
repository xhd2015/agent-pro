# Scenario

**Feature**: skip `grok update` on the chain; hit the real grok further up

```
# grok 4242 (real runner)
#   grok update 4300
#     bash 5000
#       start 6000
FindAncestorGrok(6000) -> 4242 (not 4300)
ResolveFromAncestors -> SessionID from 4242 Lsof
```

## Preconditions

- `IsGrokRunner("grok update")` is false.
- Real grok has the session open path.

## Steps

1. Set `PID=6000`.
2. Install real grok → update → bash → start.
3. OpenFiles for 4242 (update may even look session-like; ignore).

## Context

- Exclusion is classification, not missing files.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	req.Procs = []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: "/usr/local/bin/grok"},
		{PID: pidUpdate, PPID: pidGrok, Cmd: "/usr/local/bin/grok update"},
		{PID: pidBash, PPID: pidUpdate, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
	}
	req.OpenFiles = map[int][]string{
		pidGrok:   {grokSessionPath(fixtureGrokSessionID)},
		pidUpdate: {grokSessionPath(fixtureDecoyGrokSessionID)},
	}
	return nil
}
```
