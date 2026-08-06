# Scenario

**Feature**: FormatStatusText renders state, file flag, and pid+name lines

```
running fixture (file+pid)
-> Status -> FormatStatusText
-> human text with State / File / PID lines
```

## Preconditions

- Same dual-signal running fixture as `running/`.
- `req.Format = "text"`.

## Steps

1. Seed file-active session + one grok hard hit.
2. Set Format to text.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status format-text fixture")
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	req.Procs = []FixtureProc{
		{PID: 5001, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		5001: {grokOpenPath(req.SessionID)},
	}
	req.Format = "text"
	return nil
}
```
