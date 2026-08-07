# Scenario

**Feature**: `TotalBudgetRunes` drops oldest messages from the formatted block

```
msgs + TotalBudgetRunes
  -> drop oldest until Text rune length ≤ budget
  -> always prefer keeping the last (trigger) message
```

## Preconditions

- Budget applies to the **full formatted block** (header + lines), not bodies alone.
- Applied after MaxMessages (if any) and after per-message body caps.
- If only the last message remains and still exceeds budget, keep it.
- Leaves use **text-only** short messages so sizes are easy to reason about.

## Steps

1. Branch Setup clears MaxMessages so only TotalBudgetRunes varies.
2. Leaf sets messages and TotalBudgetRunes.
3. Assert Shown/OldestDropped and presence/absence of bodies.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// total-budget branch: no MaxMessages count-cap; leaves set TotalBudgetRunes.
	req.Opts = msgfmt.Options{
		MaxPerMessageRunes: 0,
		MaxMessages:        0,
		TotalBudgetRunes:   0, // leaf overrides with a positive budget
	}
	return nil
}
```
