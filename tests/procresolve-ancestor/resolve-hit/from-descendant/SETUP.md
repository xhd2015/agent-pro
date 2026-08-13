# Scenario

**Feature**: start at grok-fork under bash under grok; session from grok Lsof

```
# grok 4242 --resume <wrong> --session-id <wrong>
#   bash 5000
#     grok-fork 6000   <- start
Lsof(4242) -> …/.grok/sessions/…/019fabcdef-…/…
FindAncestorGrok(6000) -> pid 4242
ResolveFromAncestors(6000) -> Kind=grok, SessionID=019fabcdef-…
```

## Preconditions

- Start pid is **not** a grok runner (today’s descendant walk would miss the parent).
- Grok cmdline flags must not be the session source.

## Steps

1. Set `PID=6000`.
2. Install grok→bash→start.
3. OpenFiles only for 4242.

## Context

- A descendant-only `ResolveFromPID(6000)` would return `Kind=none`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidStart
	req.Procs = []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: grokCmdWithIgnoredFlags()},
		{PID: pidBash, PPID: pidGrok, Cmd: "/bin/bash"},
		{PID: pidStart, PPID: pidBash, Cmd: "/usr/local/bin/grok-fork"},
	}
	req.OpenFiles = map[int][]string{
		pidGrok: {grokSessionPath(fixtureGrokSessionID)},
	}
	return nil
}
```
