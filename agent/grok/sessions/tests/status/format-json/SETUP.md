# Scenario

**Feature**: FormatStatusJSON emits locked fields without ANSI

```
running fixture
-> Status -> FormatStatusJSON
-> JSON: session_id, state, file_active, pid_checked, pids[{pid,name,cmd}]
```

## Preconditions

- Running dual-signal fixture.
- `req.Format = "json"`.

## Steps

1. Seed file-active + one grok hard hit.
2. Set Format to json.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status format-json fixture")
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	req.Procs = []FixtureProc{
		{PID: 5001, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		5001: {grokOpenPath(req.SessionID)},
	}
	req.Format = "json"
	return nil
}
```
