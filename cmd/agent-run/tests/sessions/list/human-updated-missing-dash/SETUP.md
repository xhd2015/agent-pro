# Scenario

**Feature**: human list shows `-` when session has no updated_at or created_at

```
seed meta without timestamps -> sessions --limit 0
  -> UPDATED cell is "-"
```

## Preconditions

- Prefer `updated_at`; else `created_at`; both empty → `-`.
- JSON path is not exercised here (see `json-updated-absolute`).

## Steps

1. Write `meta.json` with runner/status only (no created_at/updated_at).
2. Run human list with `--limit 0`.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := filepath.Join(req.Home, "sessions", "no_times")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	meta := map[string]any{
		"runner":     "fake-codex",
		"session_id": "no_times",
		"status":     "finished",
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
	req.Args = append(req.Args, "--limit", "0")
	return nil
}
```
