# Scenario

**Feature**: explicit --or with multi --grep matches default OR keep-set

```
# same alpha/beta/gamma store; explain list --grep alpha --grep beta --or
-> alpha + beta kept; gamma absent (same as default OR)
```

## Preconditions

- Same three exclusive sessions as multi-default-or.

## Steps

1. Seed alpha / beta / gamma.
2. Args: `list --grep alpha --grep beta --or`.
3. Assert same keep-set as default OR.

## Context

- `--or` is explicit OR; must not change semantics vs omitting the flag.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list", "--grep", "alpha", "--grep", "beta", "--or"}
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
