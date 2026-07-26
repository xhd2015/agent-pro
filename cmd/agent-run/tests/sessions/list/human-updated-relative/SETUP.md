# Scenario

**Feature**: human sessions list UPDATED column uses relative age strings

```
seed sessions with known ages from wall clock
  -> agent-run sessions --limit 0
  -> UPDATED cells: "1h2m ago", "1h ago", "4d5h ago", "4d ago", "90d ago"
  # not RFC3339; exact strings use stable multi-minute ages (no sub-second race)
```

## Preconditions

- Human list mode (no `--json`).
- Session ages are large enough that a few seconds of list latency cannot change the formatted string.
- Pure-function exact strings (including `just now` / `2s ago`) live under
  `pkgs/agentstorage/tests/reltime/`.

## Steps

1. Seed five sessions with distinct ages and ids that sort predictably by updated_at.
2. Run `agent-run sessions --limit 0`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	now := time.Now().UTC()
	// Ages chosen so small wall-clock drift during the test cannot change the
	// relative label (minutes-level headroom inside each unit boundary).
	seedFlatSessionMeta(t, req.Home, "rel_1h2m", "fake-codex", "finished",
		now.Add(-(1*time.Hour + 2*time.Minute + 30*time.Second)).Format(time.RFC3339))
	seedFlatSessionMeta(t, req.Home, "rel_1h", "fake-codex", "finished",
		now.Add(-(1*time.Hour + 5*time.Second)).Format(time.RFC3339))
	seedFlatSessionMeta(t, req.Home, "rel_4d5h", "fake-opencode", "finished",
		now.Add(-(4*24*time.Hour + 5*time.Hour + 12*time.Minute)).Format(time.RFC3339))
	seedFlatSessionMeta(t, req.Home, "rel_4d", "fake-opencode", "finished",
		now.Add(-(4*24*time.Hour + 2*time.Minute)).Format(time.RFC3339))
	seedFlatSessionMeta(t, req.Home, "rel_90d", "fake-codex", "finished",
		now.Add(-(90 * 24 * time.Hour)).Format(time.RFC3339))
	req.Args = append(req.Args, "--limit", "0")
	return nil
}
```
