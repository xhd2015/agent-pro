# Scenario

**Feature**: modern Grok TUI post-turn idle chrome is open-ready and sendable

```
snapshot: "Turn completed" + idle ❯ + "Grok 4.5" + "Shift+Tab:mode"
  -> CheckWritable: ready=true, state=idle
  -> BannerDetected(legacy) = false
  -> OpenReady = true, screen_class = idle
```

Post-turn modern idle is the primary success shape for `--open` after attach when no legacy
banner markers are present.

## Preconditions

- Fixture `grok-modern-idle-input-post-turn.txt` copied from
  `script/debug/grok-screen-snapshots/04-idle-input-ready.txt`.

## Steps

1. Set `req.FixtureFile` to modern idle fixture.

## Context

- M3: desired open-ready for modern idle (RED on OpenReady/Classify until implementer).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureModernIdle
	return nil
}
```
