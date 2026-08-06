# Scenario

**Feature**: ArchivePath ending in .tar.gz creates archive and keeps dir

```
seedStandardWorld
  + OutDir empty (temp dir)
  + ArchivePath = {temp}/session-backup.tar.gz (must not exist)
  -> Backup
  -> temp Dir kept; ArchivePath file created; Result.ArchivePath set
```

## Preconditions

- Archive path ends with `.tar.gz` and does not exist before Backup.
- OutDir empty so Backup uses a generated temp dir that is **kept**.

## Steps

1. Set `ArchivePath` under `req.TempDir`.
2. Leave `OutDir` empty.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = ""
	req.ArchivePath = filepath.Join(req.TempDir, "session-backup.tar.gz")
	return nil
}
```
