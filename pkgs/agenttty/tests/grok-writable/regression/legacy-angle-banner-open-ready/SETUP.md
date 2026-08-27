# Scenario

**Feature**: legacy `Grok ›` remains open-ready via banner markers; writable is section-only

```
snapshot: "Grok › prompt" + "Response: hi"
  -> CheckWritable: ready=false, state=unknown (no boxed composer section)
  -> BannerDetected(legacy) = true
  -> OpenReady = true
```

Fake TUI and historical fixtures use `Grok ›` / `GROK_TTY_BANNER`. Open-ready must keep these
paths green for `waitForBannerRemote` / `OpenReady`. Writable idle no longer treats bare `Grok ›`
as a prompt.

## Preconditions

- Fixture `grok-regression-idle-legacy-angle-response.txt` (existing seed; same as writable_test legacy).

## Steps

1. Set `req.FixtureFile` to legacy angle-response fixture.

## Context

- L1: documents that legacy markers still count as open-ready (compat), while writable is section-based only.

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
