# Scenario

**Feature**: status is inactive when neither file-active nor live pid

```
writeStatusSession + writeActiveSessions() // empty list
  + no live hits
-> Status(..., checkPID=true)
-> State=inactive, FileActive=false, PIDs empty
```

## Preconditions

- Session directory exists (Find succeeds).
- `active_sessions.json` has empty `sessions` array (or id not listed).
- No matching live PIDs.

## Steps

1. Seed session without listing it as active.
2. Empty live injectables.
3. Run Status with PID check.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status inactive fixture")
	writeActiveSessions(t, req.GrokHome /* none */)
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	return nil
}
```
