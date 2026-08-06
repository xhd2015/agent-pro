# Scenario

**Feature**: file-active session aborts backup (either-signal busy gate)

```
seed parent session + active_sessions lists id
  + LivePIDs empty
  -> Backup
  -> error; no payload under OutDir
```

## Preconditions

- Session exists on disk.
- `active_sessions.json` lists the parent id → `IsFileActive` true.
- No live PIDs injected.
- Either-signal gate: file-active alone is enough to error.

## Steps

1. Seed parent session (no children required).
2. Mark file-active.
3. Set OutDir for no-payload assert.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ws := filepath.Join(req.TempDir, "ws-active")
	mustMkdir(t, ws)
	req.CWD = absPath(t, ws)
	req.CWDKey = encodeCWD(t, req.CWD)
	req.SessionID = fixtureBackupParentID
	writeSessionDir(t, req.GrokHome, req.SessionID, req.CWD, "active parent", "ACTIVE-PARENT")
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	req.ExpectActive = true
	req.OutDir = filepath.Join(req.TempDir, "backup-file-active-out")
	// Explicit empty live list.
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	return nil
}
```
