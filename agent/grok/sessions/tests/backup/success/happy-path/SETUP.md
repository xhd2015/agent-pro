# Scenario

**Feature**: full backup of parent + one child to a generated temp dir

```
# inactive parent+child, prompt lines, relocation lock, logs, sqlite marker
seedStandardWorld (OutDir empty)
  -> Backup(parent, IncludeChildren=true, Live empty)
  -> temp Dir: manifest.json + payload sessions + prompt_history.session.jsonl
     + bookkeeping/relocations/<id>.lock
```

## Preconditions

- Neither `OutDir` nor `ArchivePath` set → implementation creates a temp dir.
- Include children (default).
- Not file-active; no live PIDs.

## Steps

1. Use standard world from success grouping.
2. Leave output paths empty so Backup allocates a temp directory.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// seedStandardWorld already ran in success/Setup.
	req.OutDir = ""
	req.ArchivePath = ""
	req.NoChildren = false
	return nil
}
```
