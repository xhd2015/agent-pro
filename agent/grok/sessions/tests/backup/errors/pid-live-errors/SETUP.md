# Scenario

**Feature**: live grok PID hard-hit aborts backup (either-signal busy gate)

```
seed parent (inactive in active_sessions)
  + inject ListProcs/Lsof with matching grok runner
  -> Backup
  -> error; no payload under OutDir
```

## Preconditions

- Session exists; **not** listed in `active_sessions.json`.
- Injected process is a grok runner with open-file hard hit on session id.
- Either-signal gate: live PID alone is enough to error.

## Steps

1. Seed parent session; empty active list.
2. Inject one matching live PID.
3. Set OutDir.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ws := filepath.Join(req.TempDir, "ws-live")
	mustMkdir(t, ws)
	req.CWD = absPath(t, ws)
	req.CWDKey = encodeCWD(t, req.CWD)
	req.SessionID = fixtureBackupParentID
	writeSessionDir(t, req.GrokHome, req.SessionID, req.CWD, "live parent", "LIVE-PARENT")
	writeActiveSessions(t, req.GrokHome /* none */)

	req.Procs = []FixtureProc{
		{PID: 6101, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		6101: {grokOpenPath(req.SessionID)},
	}
	req.OutDir = filepath.Join(req.TempDir, "backup-pid-live-out")
	return nil
}
```
