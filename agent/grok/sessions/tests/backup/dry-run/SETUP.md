# Scenario

**Feature**: dry-run plans a backup without writing any artifact

```
# inactive parent+child world; DryRun=true
seedStandardWorld -> sessions.Backup(..., DryRun=true)
  -> BackupResult{DryRun, PlannedFiles, PlannedBytes, RelatedSessions}
  -> write nothing (no OutDir / archive / manifest)
```

## Preconditions

- `BackupOptions.DryRun=true` (Classic TDD RED until implementer).
- Session not file-active; no live PIDs (success leaves).
- Busy gate and archive suffix validation still apply.
- Leaves may set `OutDir` / `ArchivePath` / `NoChildren` / error conditions.

## Steps

1. Seed standard parent+child world.
2. Set `req.DryRun = true`.
3. Leaf adjusts OutDir, children, or busy/error fixtures.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedStandardWorld(t, req)
	req.DryRun = true
	req.NoChildren = false
	return nil
}
```
