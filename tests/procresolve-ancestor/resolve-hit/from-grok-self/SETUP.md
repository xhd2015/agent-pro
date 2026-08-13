# Scenario

**Feature**: start pid is itself the grok runner

```
# single-node / self-first walk
start=4242 Cmd=/usr/local/bin/grok
Lsof(4242) -> …/.grok/sessions/…/019fabcdef-…/…
FindAncestorGrok(4242) -> pid 4242
ResolveFromAncestors(4242) -> Kind=grok, SessionID=019fabcdef-…
```

## Preconditions

- Walk includes startPID before PPID, so a grok start wins immediately.

## Steps

1. Set `PID=4242`.
2. Install the grok (and a non-grok parent init).
3. OpenFiles for 4242.

## Context

- `FindAncestorGrok` must report the start pid, not walk past it.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidGrok
	req.Procs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "/sbin/init"},
		{PID: pidGrok, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		pidGrok: {grokSessionPath(fixtureGrokSessionID)},
	}
	return nil
}
```
