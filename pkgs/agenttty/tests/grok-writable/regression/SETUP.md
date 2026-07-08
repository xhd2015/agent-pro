# Scenario

**Bug**: targeted regression leaves for known writable-detection edge cases

```
load single regression fixture
  -> CheckWritable
  -> assert specific ready/state/reason
```

## Preconditions

- Regression fixtures live in shared `pkgs/agenttty/testdata/grok-writable/`.
- `git-working-tree-idle-prompt` expects post-fix behavior (`ready=true`); RED on current code.

## Steps

1. Each leaf `Setup` sets `req.FixtureFile` to its regression fixture basename.

## Context

- F2–F4 are narrow guards complementing the full fixture table (F1).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunAllFixtures = false
	return nil
}
```