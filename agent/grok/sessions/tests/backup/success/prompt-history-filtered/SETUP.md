# Scenario

**Feature**: prompt_history extract contains only parent+child session_id lines

```
seedStandardWorld (4 prompt lines: 2 parent, 1 child, 1 noise)
  -> Backup
  -> payload/.../prompt_history.session.jsonl has 3 lines; no noise id
```

## Preconditions

- Shared `prompt_history.jsonl` contains lines for parent, child, and a noise id.
- Filtered extract file name: `prompt_history.session.jsonl` under payload sessions cwd key.

## Steps

1. Use standard world (already seeds mixed prompt lines).
2. Set explicit OutDir for stable assert path.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.OutDir = filepath.Join(req.TempDir, "backup-prompt")
	return nil
}
```
