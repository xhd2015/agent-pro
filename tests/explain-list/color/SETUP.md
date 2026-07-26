# Scenario

**Feature**: color policy for explain list cards

```
# priority: --color > NO_COLOR > TTY auto
explain list [--color] (+ optional NO_COLOR) -> ANSI on or off per policy
# harness stdout is a pipe (non-TTY) so auto is off without --color
```

## Preconditions

- Shared one-session fixture for all color leaves (short Q/A).
- Root `Run` strips parent `NO_COLOR` / `FORCE_COLOR` / `CLICOLOR_FORCE`.

## Steps

1. Grouping seeds nothing; leaves set Args/EnvExtra and the shared session.
2. Assert presence/absence of specific SGR sequences.

## Context

- When color on: Q=`\x1b[1;36m`, A=`\x1b[1;32m`, meta/dim=`\x1b[2m`, reset `\x1b[0m`.
- Bodies remain uncolored.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("color setup: explain binary not built")
	}
	return nil
}

// colorFixtureSession is a short one-turn session used by all color leaves.
func colorFixtureSession() SessionSeed {
	return simpleSession(
		"2026-07-13-14-30-05-colorfix-aaaaaaaa",
		"opencode", "deepseek-chat",
		"color q", "color a",
	)
}
```
