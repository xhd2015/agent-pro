# Scenario

**Feature**: dry-run still errors when session is file-active (busy gate)

```
seed parent listed in active_sessions
  + DryRun=true + empty LivePIDs
  -> Backup -> error; nothing written under OutDir
```

## Preconditions

- Session exists on disk.
- `active_sessions.json` lists the parent id → `IsFileActive` true.
- Dry-run does **not** bypass the busy gate.
- No live PIDs injected.

## Steps

1. Re-mark parent as file-active (grouping already seeded world).
2. Set OutDir for no-write assert.
3. Keep DryRun true.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeActiveSessions(t, req.GrokHome, req.SessionID)
	req.ExpectActive = true
	req.OutDir = filepath.Join(req.TempDir, "dry-run-file-active-out")
	req.ArchivePath = ""
	req.Procs = nil
	req.OpenFiles = map[int][]string{}
	return nil
}
```
