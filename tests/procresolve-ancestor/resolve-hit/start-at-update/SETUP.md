# Scenario

**Feature**: start pid is `grok update`; walk continues to the parent real grok

```
# grok 4242
#   grok update 4300  <- start
FindAncestorGrok(4300) -> 4242
ResolveFromAncestors(4300) -> SessionID from 4242
```

## Preconditions

- Start itself is not `IsGrokRunner`.
- Parent is a real grok with a session path.

## Steps

1. Set `PID=4300`.
2. Install real grok parent + update start.
3. OpenFiles for 4242.

## Context

- Walk is start **then** PPID; a non-runner start is not a hit.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = pidUpdate
	req.Procs = []FixtureProc{
		{PID: pidGrok, PPID: 1, Cmd: "/usr/local/bin/grok"},
		{PID: pidUpdate, PPID: pidGrok, Cmd: "/usr/local/bin/grok update"},
	}
	req.OpenFiles = map[int][]string{
		pidGrok: {grokSessionPath(fixtureGrokSessionID)},
	}
	return nil
}
```
