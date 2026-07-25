# Scenario

**Feature**: input pid is itself a grok runner with session open files

```
# single-node tree
input pid=100 Cmd=/usr/local/bin/grok
  Lsof(100) -> …/.grok/sessions/…/019f…/events.jsonl
ResolveFromPID(100) -> Kind=grok, Source=open-files, RunnerPID=100
```

## Preconditions

- Snapshot contains only the grok process (pid 100, ppid 1).
- No cmdline session flags are required; path alone supplies the uuid.

## Steps

1. Set `PID=100`.
2. Install one `FixtureProc` for grok.
3. Map Lsof for 100 to `grokSessionPath(fixtureGrokSessionID)`.

## Context

- Source must be exactly `open-files` (input is the runner; no tree descent needed for candidate).
- Tree should include the input node with role `input` or `grok` (implementer may set Role=`input` for the input pid even when it is also a runner; runner classification still yields Kind=grok). Prefer: input node Role is `input` when reporting the requested root, and Kind/Runner still reflect grok. If implementer uses Role=`grok` on the single node, tests accept either `input` or `grok` for pid 100.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = 100
	req.Procs = []FixtureProc{
		{PID: 100, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		100: {grokSessionPath(fixtureGrokSessionID)},
	}
	return nil
}
```
