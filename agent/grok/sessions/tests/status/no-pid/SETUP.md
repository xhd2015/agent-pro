# Scenario

**Feature**: --no-pid skips live scan; state comes from file-active only

```
writeStatusSession + writeActiveSessions(id)
  + injectable grok pid that WOULD match
  + NoPID / checkPID=false
-> Status
-> PIDChecked=false, PIDs empty, State=marked-active (file only)
```

## Preconditions

- Session is file-active.
- Injectables include a hard-hit grok pid that must be **ignored** when
  `checkPID` is false.

## Steps

1. Seed session and mark file-active.
2. Inject matching grok open-file hit (would produce running if checked).
3. Set `req.NoPID = true`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status no-pid fixture")
	writeActiveSessions(t, req.GrokHome, req.SessionID)

	// Would match if PID check were on — must be ignored.
	req.Procs = []FixtureProc{
		{PID: 6001, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		6001: {grokOpenPath(req.SessionID)},
	}
	req.NoPID = true
	return nil
}
```
