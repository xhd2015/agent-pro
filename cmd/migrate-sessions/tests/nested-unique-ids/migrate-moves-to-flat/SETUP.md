# Scenario

**Feature**: nested unique ids migrate to flat layout with backup and `.layout` v2

```
seed nested two runners -> migrate-sessions --home H
-> flat sessions/<id>/; events preserved; backup under H/backups/; .layout version 2
```

## Preconditions

- Nested: `fake-codex/sess_a` and `fake-opencode/sess_b` with distinct events.
- Also seed a non-migrated `fake-codex-registry/` sibling to prove registry untouched.

## Steps

1. Write nested sessions + a dummy registry dir.
2. Run migrator (default backup dir).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeNestedSession(t, req.Home, "fake-codex", "sess_a", "finished", "2026-07-01T10:00:00Z", "event-a")
	writeNestedSession(t, req.Home, "fake-opencode", "sess_b", "running", "2026-07-01T11:00:00Z", "event-b")
	// non-migrated tree
	reg := filepath.Join(req.Home, "fake-codex-registry", "live")
	if err := os.MkdirAll(reg, 0o755); err != nil {
		t.Fatalf("mkdir registry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reg, "marker.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write registry marker: %v", err)
	}
	req.Args = []string{"--home", req.Home}
	return nil
}
```
