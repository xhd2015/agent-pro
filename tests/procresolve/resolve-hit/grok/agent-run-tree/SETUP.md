# Scenario

**Feature**: agent-run → serve → grok grandchild; session from open files on grok

```
# three-level tree
pid 200 agent-run run …     (input)
  -> pid 201 agent-run serve …
       -> pid 202 /usr/local/bin/grok
            Lsof(202) -> …/.grok/sessions/…/019f…/…
ResolveFromPID(200) -> Kind=grok, Source=open-files+tree, RunnerPID=202
```

## Preconditions

- Input is agent-run (not the runner itself).
- Child serve and grandchild grok exist as descendants.
- Only the grok pid has a session open path; agent-run nodes have empty Lsof.

## Steps

1. Set `PID=200`.
2. Install three procs: 200→201→202 with cmds agent-run run / agent-run serve / grok.
3. OpenFiles only for 202.

## Context

- Tree must report 3 nodes with roles:
  - 200: `input` (or `agent-run` if dual-tagged — require Role `input` for the requested root)
  - 201: `agent-run-serve`
  - 202: `grok`
- Source must be `open-files+tree` (tree walk found the runner under input).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = 200
	req.Procs = []FixtureProc{
		{PID: 200, PPID: 1, Cmd: "/usr/local/bin/agent-run run --session-id=ignored-cli"},
		{PID: 201, PPID: 200, Cmd: "/usr/local/bin/agent-run serve --session-id=ignored-cli"},
		{PID: 202, PPID: 201, Cmd: "/usr/local/bin/grok"},
	}
	// Intentionally put --session-id on agent-run cmdline: must NOT be primary source.
	req.OpenFiles = map[int][]string{
		202: {grokSessionPath(fixtureGrokSessionID)},
	}
	return nil
}
```
