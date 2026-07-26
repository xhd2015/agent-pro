# Scenario

**Feature**: dry-run on nested home leaves tree intact

```
seed nested -> migrate-sessions --home H --dry-run -> exit 0; nested paths remain; no flat move
```

## Preconditions

- At least one nested session.

## Steps

1. Seed nested session `fake-codex/dry_sess`.
2. Run with `--dry-run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	writeNestedSession(t, req.Home, "fake-codex", "dry_sess", "finished", "2026-07-03T00:00:00Z", "dry-event")
	req.Args = []string{"--home", req.Home, "--dry-run"}
	return nil
}
```
