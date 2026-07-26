# Scenario

**Feature**: `--json` list keeps absolute timestamps (not relative human ages)

```
seed session with known updated_at RFC3339
  -> agent-run sessions --json --limit 0
  -> sessions[i].updated_at is the absolute string; no "ago" / "just now"
```

## Preconditions

- JSON schema unchanged: absolute `created_at` / `updated_at` strings.
- Human relative formatting must not leak into JSON fields.

## Steps

1. Seed one session with fixed absolute timestamps.
2. Run `agent-run sessions --json --limit 0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	seedFlatSessionMeta(t, req.Home, "json_abs", "fake-codex", "finished", "2026-07-01T15:04:05Z")
	req.Args = append(req.Args, "--json", "--limit", "0")
	return nil
}
```
