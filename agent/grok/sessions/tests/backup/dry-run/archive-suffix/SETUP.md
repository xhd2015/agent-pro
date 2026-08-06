# Scenario

**Feature**: dry-run still rejects ArchivePath that does not end with .tar.gz

```
seedStandardWorld + DryRun
  + ArchivePath = {temp}/backup.tgz  (wrong suffix)
  -> Backup -> error; no archive; no OutDir payload
```

## Preconditions

- Session exists, inactive, no live pids.
- `ArchivePath` does **not** end with `.tar.gz`.
- Suffix validation applies even when `DryRun=true`.

## Steps

1. Use dry-run standard world.
2. Set invalid archive suffix and OutDir.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "dry-run-suffix-out")
	req.ArchivePath = filepath.Join(req.TempDir, "backup.tgz") // not .tar.gz
	return nil
}
```
