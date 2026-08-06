# Scenario

**Feature**: dry-run with IncludeChildren=false omits child from plan

```
seedStandardWorld (child exists on disk + linked)
  -> Backup(DryRun=true, IncludeChildren=false)
  -> RelatedSessions has parent only; child not counted as related
```

## Preconditions

- Child session directory exists under grok home.
- `NoChildren` true → `IncludeChildren` pointer false.
- Dry-run mode.

## Steps

1. Use dry-run standard world.
2. Set `req.NoChildren = true`.
3. Set OutDir for no-write assert.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.NoChildren = true
	req.OutDir = filepath.Join(req.TempDir, "dry-run-no-children-out")
	req.ArchivePath = ""
	return nil
}
```
