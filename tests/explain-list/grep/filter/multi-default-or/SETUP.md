# Scenario

**Feature**: multiple --grep combine with default OR

```
# sessions: alpha-only, beta-only, gamma-only
# explain list --grep alpha --grep beta
-> alpha + beta kept; gamma absent; title 2 shown of 2
```

## Preconditions

- Three sessions with mutually exclusive markers.

## Steps

1. Seed alpha / beta / gamma sessions (beta newest).
2. Args: `list --grep alpha --grep beta` (no `--or`/`--and`).
3. Assert alpha+beta present, gamma absent, newest-first among matches.

## Context

- Default combine is OR when multiple greps are given.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "alpha", "--grep", "beta"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-01-08-00-00-gamma-gggggggg",
			"opencode", "deepseek-chat",
			"gamma only marker-gamma", "a",
		),
		simpleSession(
			"2026-07-02-08-00-00-alpha-aaaaaaaa",
			"opencode", "deepseek-chat",
			"alpha only marker-alpha", "a",
		),
		simpleSession(
			"2026-07-03-08-00-00-beta-bbbbbbbb",
			"opencode", "deepseek-chat",
			"beta only marker-beta", "a",
		),
	}
	return nil
}
```
