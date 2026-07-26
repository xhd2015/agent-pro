# Scenario

**Feature**: legacy `Grok ›` banner frames remain open-ready (backward compatibility)

```
snapshot: "Grok › prompt" + "Response: hi"
  -> CheckWritable: ready=true, state=idle
  -> BannerDetected(legacy) = true
  -> OpenReady = true
```

Fake TUI and historical fixtures use `Grok ›` / `GROK_TTY_BANNER`. Open-ready must keep these
paths green when implementer switches `waitForBannerRemote` to `OpenReady`.

## Preconditions

- Fixture `grok-regression-idle-legacy-angle-response.txt` (existing seed; same as writable_test legacy).

## Steps

1. Set `req.FixtureFile` to legacy angle-response fixture.

## Context

- L1: documents that legacy markers still count as open-ready (compat, not a regression of markers).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FixtureFile = fixtureLegacyAngleResponse
	return nil
}
```
