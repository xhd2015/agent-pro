# Scenario

**Feature**: modern Grok TUI busy/thinking chrome remains open-ready but not sendable

```
snapshot: Tasks / "Thinking…" + ❯ + "Grok 4.5" chrome
  -> CheckWritable: ready=false, state=busy, reason=agent still responding
  -> BannerDetected(legacy) = false
  -> OpenReady = true, screen_class = busy
```

Busy frames still show modern TUI chrome; open lifecycle may proceed past banner wait while
send-queue correctly blocks injection (`ready=false`).

## Preconditions

- Fixture `grok-modern-busy-thinking-tasks.txt` copied from
  `script/debug/grok-screen-snapshots/03-busy-working.txt`.

## Steps

1. Set `req.FixtureFile` to modern busy fixture.

## Context

- M2: open-ready is orthogonal to writable busy; RED on OpenReady/Classify until implementer.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FixtureFile = fixtureModernBusy
	return nil
}
```
