# Scenario

**Feature**: plain bash pid with empty Lsof yields Kind=none

```
pid 400 /bin/bash
  Lsof(400) -> []
ResolveFromPID(400) -> Kind=none, SessionID="", err=nil
```

## Preconditions

- Single bash process; not classified as grok/codex.

## Steps

1. Set `PID=400`, one bash proc, empty OpenFiles.

## Context

- Tree may still list the input node with role `input` / `other`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = 400
	req.Procs = []FixtureProc{
		{PID: 400, PPID: 1, Cmd: "/bin/bash"},
	}
	req.OpenFiles = map[int][]string{}
	return nil
}
```
