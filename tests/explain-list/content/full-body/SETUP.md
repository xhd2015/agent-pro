# Scenario

**Feature**: long message bodies print in full (no soft-truncate, no ellipsis)

```
# assistant message of 200 'x' runes
explain list -> A body is the full 200 x string; no …; no rune cap
```

## Preconditions

- One session; answer is 200 ASCII `x` characters (1 rune each).

## Steps

1. Seed session with long answer (`strings.Repeat("x", 200)`).
2. Run list.
3. Assert full body present; no ellipsis; exact card template.

## Context

- Classic TDD: current product still soft-truncates ~140 + `…` → this leaf is RED until implement.
- Plain (no `--color`); truncate/collapse removed for list bodies.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"list"}
	long := strings.Repeat("x", 200)
	req.Sessions = []SessionSeed{
		simpleSession(
			"2026-07-13-10-00-00-longmsg-cccccccc",
			"opencode", "deepseek-chat",
			"short q",
			long,
		),
	}
	return nil
}
```
