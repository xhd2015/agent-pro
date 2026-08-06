# Scenario

**Feature**: ArchivePath must not already exist

```
seed parent
  + ArchivePath = existing empty file ending .tar.gz
  -> Backup -> error; existing file left alone; no OutDir payload
```

## Preconditions

- Session exists, inactive, no live pids.
- `ArchivePath` ends with `.tar.gz` and **already exists**.

## Steps

1. Seed parent.
2. Create placeholder archive file.
3. Set OutDir.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ws := filepath.Join(req.TempDir, "ws-archive-exists")
	mustMkdir(t, ws)
	req.CWD = absPath(t, ws)
	req.CWDKey = encodeCWD(t, req.CWD)
	req.SessionID = fixtureBackupParentID
	writeSessionDir(t, req.GrokHome, req.SessionID, req.CWD, "exists parent", "EXISTS")
	writeActiveSessions(t, req.GrokHome /* none */)
	req.OutDir = filepath.Join(req.TempDir, "backup-exists-out")
	req.ArchivePath = filepath.Join(req.TempDir, "already.tar.gz")
	mustWriteFile(t, req.ArchivePath, "pre-existing-archive-bytes\n")
	return nil
}
```
