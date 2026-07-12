# Scenario

**Bug**: targeted regression leaves for writable-detection and open-lifecycle edge cases

```
load single regression fixture
  -> CheckWritable (Run)
  -> Assert: ready/state/reason; open-ready leaves also call exported APIs
```

## Preconditions

- Regression fixtures live in shared `pkgs/agenttty/testdata/grok-writable/`.
- `git-working-tree-idle-prompt` expects post-fix behavior (`ready=true`); RED on current code.
- Modern open-ready leaves (M1–M3) and enriched modal/empty/legacy leaves call exported
  open-ready APIs from Assert (RED until implementer).

## Steps

1. Clear table mode (`RunAllFixtures=false`).
2. Each leaf `Setup` sets `req.FixtureFile` to its regression fixture basename.

## Context

- F2–F4, W2, M1–M3, L1 are narrow guards complementing the full fixture table (F1).
- Writable option A for modern starting remains `ready=true`/`state=idle` (not forced loading).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunAllFixtures = false
	return nil
}
```
