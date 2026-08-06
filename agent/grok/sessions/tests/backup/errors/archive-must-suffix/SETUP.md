# Scenario

**Feature**: ArchivePath must end with .tar.gz

```
seed parent
  + ArchivePath = {temp}/backup.tgz  (wrong suffix)
  -> Backup -> error; no archive; no payload at OutDir if set
```

## Preconditions

- Session exists and is inactive / no live pids.
- `ArchivePath` does **not** end with `.tar.gz`.

## Steps

1. Seed minimal parent.
2. Set invalid archive suffix.
3. Set OutDir for no-payload assert.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ws := filepath.Join(req.TempDir, "ws-archive-suffix")
	mustMkdir(t, ws)
	req.CWD = absPath(t, ws)
	req.CWDKey = encodeCWD(t, req.CWD)
	req.SessionID = fixtureBackupParentID
	writeSessionDir(t, req.GrokHome, req.SessionID, req.CWD, "suffix parent", "SUFFIX")
	writeActiveSessions(t, req.GrokHome /* none */)
	req.OutDir = filepath.Join(req.TempDir, "backup-suffix-out")
	req.ArchivePath = filepath.Join(req.TempDir, "backup.tgz") // not .tar.gz
	return nil
}
```
