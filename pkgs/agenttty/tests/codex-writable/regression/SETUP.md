# Scenario

**Bug**: targeted regression leaves for Codex status-related writable-detection edge cases

```
load single regression fixture
  -> CheckWritable
  -> assert specific ready/state/reason
```

## Preconditions

- Regression fixtures live in shared `pkgs/agenttty/testdata/codex-writable/`.
- `update-modal-not-idle` expects post-fix behavior (`ready=false`, non-idle); RED on current code.

## Steps

1. Each leaf `Setup` sets `req.FixtureFile` to its regression fixture basename.

## Context

- F2–F8 are narrow guards complementing the full fixture table (F1).
- F4 keeps legacy `›` idle GREEN; F6–F7 lock Codex 0.146 `»` idle (RED before implementer).
- F8 locks post-turn idle when historical `• Working` remains above bottom `›` (RED before fix).
- F9 locks inject-ready (`BannerDetected`) false on live `Starting MCP servers`.
- F10 locks live `• Working` above a placeholder composer `›` as busy.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RunAllFixtures = false
	return nil
}
```
