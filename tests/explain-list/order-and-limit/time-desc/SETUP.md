# Scenario

**Feature**: newer session appears before older session

```
# older 2026-07-12 ... + newer 2026-07-13 ...
explain list -> card 1 is newer, card 2 is older
```

## Preconditions

- Two valid sessions with distinct timestamp prefixes.

## Steps

1. Seed older then newer (order of seeding irrelevant).
2. Run default `list`.
3. Assert index-1 card is the newer timestamp / question.

## Context

- Sort key is dirname timestamp, not write order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-12-09-15-22-older-bbbbbbbb",
			"opencode", "deepseek-chat",
			"older question", "older answer",
		),
		simpleSession(
			"2026-07-13-14-30-05-newer-aaaaaaaa",
			"opencode", "deepseek-chat",
			"newer question", "newer answer",
		),
	}
	return nil
}
```
