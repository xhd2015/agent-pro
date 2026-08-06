# Scenario

**Feature**: non-grok and grok-update processes are not session live PIDs

```
ListProcs:
  bash (open session path)
  grok update (open session path)
  plain node (no path)
-> LivePIDs / Status
-> PIDs empty; State not running from these alone
```

## Preconditions

- Session exists; not file-active.
- Open-file hard hits alone are insufficient without a grok **runner** cmd.
- `grok update` is classified non-runner (same as procresolve).

## Steps

1. Seed inactive session.
2. Inject bash + `grok update` both with matching open paths.
3. Assert no live PIDs / inactive state.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status skip-non-runners fixture")
	writeActiveSessions(t, req.GrokHome /* none */)

	path := grokOpenPath(req.SessionID)
	req.Procs = []FixtureProc{
		{PID: 8001, PPID: 1, Cmd: "/bin/bash -c sleep 999"},
		{PID: 8002, PPID: 1, Cmd: "/usr/local/bin/grok update"},
		{PID: 8003, PPID: 1, Cmd: "/usr/bin/node /app/server.js"},
	}
	req.OpenFiles = map[int][]string{
		8001: {path},
		8002: {path},
		// 8003: no open files
	}
	return nil
}
```
