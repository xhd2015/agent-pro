# Scenario

**Feature**: LivePIDs returns all matching grok runners sorted by PID

```
two grok procs (pids 7002, 7001) both hard-hit same session
  (inject higher pid first in ListProcs order)
-> Status / LivePIDs
-> PIDs sorted ascending: 7001 then 7002; Name=basename(argv0)
```

## Preconditions

- Session exists (Find succeeds); file-active optional (use empty active list).
- Two distinct grok runner cmds with different basenames to prove Name mapping:
  - `/opt/bin/grok` → Name `grok`
  - `/usr/local/bin/grok-agent` would be name `grok-agent` — prefer both real
    `grok` basenames with different full Cmd paths; Name still `grok`.
  - Use cmds `/opt/homebrew/bin/grok` and `/usr/local/bin/grok` — both Name
    `grok`; assert PID order and Cmd strings.

## Steps

1. Seed inactive (file) session on disk.
2. Inject two matching pids **out of PID order** in `Procs` slice.
3. Open-file hard hits for both.
4. Op remains default Status (also validates multi PID on SessionStatus).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status multi-pid fixture")
	writeActiveSessions(t, req.GrokHome /* none */)

	// Intentionally list higher PID first — result must still sort by PID asc.
	req.Procs = []FixtureProc{
		{PID: 7002, PPID: 1, Cmd: "/usr/local/bin/grok --session-dir=x"},
		{PID: 7001, PPID: 1, Cmd: "/opt/homebrew/bin/grok"},
	}
	path := grokOpenPath(req.SessionID)
	req.OpenFiles = map[int][]string{
		7001: {path},
		7002: {path},
	}
	return nil
}
```
