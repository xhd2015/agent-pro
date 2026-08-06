# Scenario

**Feature**: status is marked-active when file-active but no live pid

```
writeStatusSession + writeActiveSessions(id)
  + ListProcs empty (or no hard hits)
-> Status(..., checkPID=true)
-> State=marked-active, FileActive=true, PIDs empty
```

## Preconditions

- Session exists and is listed in `active_sessions.json`.
- No injectable process hard-hits the session open path.

## Steps

1. Seed session and mark file-active.
2. Leave `Procs` empty and `OpenFiles` empty.
3. Check PIDs (default).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status marked-active fixture")
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	// Explicit empty live snapshot: file-active only.
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	return nil
}
```
