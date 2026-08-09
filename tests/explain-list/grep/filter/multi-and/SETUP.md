# Scenario

**Feature**: multi --grep --and requires every pattern in the session

```
# sessions: both, alpha-only, beta-only
# explain list --grep alpha --grep beta --and
-> only "both" kept; title 1 shown of 1
```

## Preconditions

- Three sessions: one containing both patterns in one body; two with only one.

## Steps

1. Seed both / alpha-only / beta-only.
2. Args: `list --grep alpha --grep beta --and`.
3. Assert only both-session markers.

## Context

- Session-level AND: a session is kept only if every pattern matches somewhere
  in its message bodies.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "alpha", "--grep", "beta", "--and"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-01-09-00-00-alphaonly-aaaaaaaa",
			"opencode", "deepseek-chat",
			"alpha only marker-alpha-only", "a",
		),
		simpleSession(
			"2026-07-02-09-00-00-betaonly-bbbbbbbb",
			"opencode", "deepseek-chat",
			"beta only marker-beta-only", "a",
		),
		simpleSession(
			"2026-07-03-09-00-00-both-cccccccc",
			"opencode", "deepseek-chat",
			"alpha and beta together marker-both", "a",
		),
	}
	return nil
}
```
