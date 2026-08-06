# Scenario

**Feature**: status is running when file-active and a live grok pid hard-hits

```
writeStatusSession + writeActiveSessions(id)
  + ListProcs: grok pid 5001
  + Lsof(5001): …/.grok/sessions/…/<id>/…
-> Status(..., checkPID=true, live)
-> State=running, FileActive=true, one LivePID
```

## Preconditions

- Session directory exists with `summary.json`.
- `active_sessions.json` lists the session id.
- Injected process is a grok runner (not `grok update`) with matching open path.

## Steps

1. Seed session under fixture cwd.
2. Mark session file-active.
3. Inject one grok process and open-file hard hit.
4. Set `SessionID`; leave `NoPID` false (check PIDs).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status running fixture")
	writeActiveSessions(t, req.GrokHome, req.SessionID)

	req.Procs = []FixtureProc{
		{PID: 5001, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		5001: {grokOpenPath(req.SessionID)},
	}
	return nil
}
```
