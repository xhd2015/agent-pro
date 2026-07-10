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

- F2–F5 are narrow guards complementing the full fixture table (F1).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunAllFixtures = false
	return nil
}
```
