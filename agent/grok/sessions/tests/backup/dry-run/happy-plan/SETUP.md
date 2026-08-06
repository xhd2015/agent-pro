# Scenario

**Feature**: dry-run happy plan includes children and does not create OutDir

```
seedStandardWorld + DryRun + OutDir set (missing on disk)
  -> Backup
  -> DryRun true; PlannedFiles > 0; RelatedSessions has parent+child
  -> OutDir never created
```

## Preconditions

- Explicit `OutDir` that does not exist yet.
- Include children (default).
- Not busy.

## Steps

1. Use dry-run grouping standard world.
2. Set `req.OutDir` under temp; leave archive empty.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "dry-run-happy-out")
	req.ArchivePath = ""
	req.NoChildren = false
	return nil
}
```
