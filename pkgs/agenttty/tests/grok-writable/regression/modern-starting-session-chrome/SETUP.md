# Scenario

**Feature**: modern Grok TUI starting-session chrome is open-ready without legacy banner markers

```
snapshot: "Starting session…" + ❯ + "Grok 4.5" + "Shift+Tab:mode"
  -> CheckWritable: option A ready=true, state=idle
  -> BannerDetected(legacy) = false
  -> OpenReady = true, screen_class = starting
```

SeaTalk / real modern grok frames show starting chrome without `GROK_TTY_BANNER` or `Grok ›`.
Open lifecycle must accept this frame; send-queue writable remains option A (idle/ready, not loading).

## Preconditions

- Fixture `grok-modern-starting-session-chrome.txt` copied from
  `script/debug/grok-screen-snapshots/02-early-tui-chrome-or-input.txt`.

## Steps

1. Set `req.FixtureFile` to modern starting fixture.

## Context

- M1: freezes modern starting for open-ready (RED on OpenReady/Classify until implementer).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureModernStarting
	return nil
}
```
