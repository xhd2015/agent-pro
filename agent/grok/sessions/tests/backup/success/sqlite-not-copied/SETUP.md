# Scenario

**Feature**: session_search.sqlite is never copied; manifest notes presence

```
seedStandardWorld (sqlite marker at sessions/session_search.sqlite)
  -> Backup
  -> no sqlite under payload; manifest.sqlite.present=true
```

## Preconditions

- Source `sessions/session_search.sqlite` exists with known marker bytes.
- Backup must not place that file under the payload tree.

## Steps

1. Use standard world.
2. Explicit OutDir.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "backup-sqlite")
	return nil
}
```
