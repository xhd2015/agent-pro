# Scenario

**Feature**: unified logs appear only as manifest meta (count + last ≤3 lines)

```
seedStandardWorld (5 matching unified.jsonl lines for parent+child)
  -> Backup
  -> manifest.logs.match_count=5; last_lines len≤3 with time; no log file under payload
```

## Preconditions

- Source `logs/unified.jsonl` has matches for parent+child ids (and noise).
- No log bytes may be copied under `payload/`.

## Steps

1. Use standard world (`LogMatchCount=5`).
2. Explicit OutDir for stable paths.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "backup-logs")
	return nil
}
```
