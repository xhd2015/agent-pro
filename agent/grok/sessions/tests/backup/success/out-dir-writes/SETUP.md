# Scenario

**Feature**: explicit OutDir receives the backup tree

```
seedStandardWorld
  + OutDir = {temp}/backup-out (empty / non-existent)
  -> Backup
  -> OutDir/manifest.json + OutDir/payload/...
  -> Result.Dir == OutDir
```

## Preconditions

- `OutDir` does not exist yet (or is empty).
- No archive.

## Steps

1. Set `req.OutDir` to a path under `req.TempDir`.
2. Leave `ArchivePath` empty.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "backup-out")
	req.ArchivePath = ""
	return nil
}
```
