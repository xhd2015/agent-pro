# Scenario

**Feature**: already-flat home exits 0 without destructive change

```
seed flat sess_x + .layout v2 -> migrate-sessions --home H -> exit 0; sess_x intact
```

## Preconditions

- Detect already-flat via `.layout` version 2 (or all top-level children are session dirs with meta).

## Steps

1. Write flat session and `.layout`.
2. Capture pre-migration meta bytes; run migrator.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeFlatSession(t, req.Home, "sess_x", "fake-codex", "finished", "2026-07-02T00:00:00Z")
	writeLayoutV2(t, req.Home)
	// marker file to detect unwanted rewrite of session tree
	marker := filepath.Join(req.Home, "sessions", "sess_x", "events.jsonl")
	if err := os.WriteFile(marker, []byte(`{"type":"message","text":"keep-me"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write events: %v", err)
	}
	req.Args = []string{"--home", req.Home}
	return nil
}
```
