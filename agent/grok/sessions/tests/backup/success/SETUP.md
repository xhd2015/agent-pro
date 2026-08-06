# Scenario

**Feature**: successful Backup writes a self-describing directory

```
# inactive session (empty LivePIDs, not file-active)
seedStandardWorld -> sessions.Backup(...)
  -> BackupResult.Dir with manifest.json + payload/
```

## Preconditions

- Session is **not** file-active.
- Injected `LiveOptions` list is empty (no live PIDs).
- Parent session exists on disk (and child unless leaf overrides).

## Steps

1. Seed a standard parent+child world unless a leaf replaces fixtures.
2. Leave `NoChildren` false by default (include children).
3. Leaf may set `OutDir` / `ArchivePath` / content variants.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Default success world; leaves may re-seed or adjust fields.
	seedStandardWorld(t, req)
	return nil
}
```
