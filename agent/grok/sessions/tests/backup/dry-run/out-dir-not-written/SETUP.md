# Scenario

**Feature**: dry-run never creates OutDir or archive even when both set

```
seedStandardWorld
  + DryRun + OutDir + ArchivePath (.tar.gz, both missing)
  -> Backup
  -> success plan; neither OutDir nor archive path exists on disk
```

## Preconditions

- Both `OutDir` and valid `ArchivePath` (`.tar.gz`) are set and do not exist yet.
- Live backup would create them; dry-run must not.

## Steps

1. Use dry-run standard world.
2. Set fresh OutDir and ArchivePath under temp.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "dry-run-never-write-out")
	req.ArchivePath = filepath.Join(req.TempDir, "dry-run-never-write.tar.gz")
	req.NoChildren = false
	return nil
}
```
