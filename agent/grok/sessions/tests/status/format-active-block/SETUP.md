# Scenario

**Feature**: FormatActiveBlock provides dual-signal Active section for session info

```
running fixture
-> Status -> FormatActiveBlock
-> text block mentioning file active + live pid (for CLI info append)
```

## Preconditions

- Running dual-signal fixture.
- `FormatInfoText` signature stays unchanged; CLI appends Active block.
- `req.Format = "active-block"`.

## Steps

1. Seed file-active session + one grok hard hit.
2. Format as active-block.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureStatusSessionID
	writeStatusSession(t, req.GrokHome, req.SessionID, fixtureStatusCWD, "status active-block fixture")
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	req.Procs = []FixtureProc{
		{PID: 5001, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		5001: {grokOpenPath(req.SessionID)},
	}
	req.Format = "active-block"
	return nil
}
```
