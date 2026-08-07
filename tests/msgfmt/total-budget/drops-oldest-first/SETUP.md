# Scenario

**Feature**: tight budget drops oldest first while keeping newer messages

```
msgs: OLD_UNIQUE, MID_UNIQUE, NEW_UNIQUE (text-only)
TotalBudgetRunes small enough that all three cannot fit
  -> oldest body absent; newest body present
  -> OldestDropped >= 1; Shown < SourceCount
```

## Preconditions

- Distinct unique body tokens so absence is unambiguous.
- Budget chosen so the full 3-message block cannot fit, but a suffix of
  messages can (implementer drops oldest until fit).
- Use `TotalBudgetRunes` equal to the rune length of the **2-newest** full
  block so exactly one oldest message is dropped when the algorithm is
  “drop until fits”.

Locked budget: set to the exact rune count of:

```text
Chat history (showing 2 of 3):
MID_UNIQUE
NEW_UNIQUE
```

(including trailing newline). That forces drop of `OLD_UNIQUE` if the
formatter recomputes the header for the surviving set.

## Steps

1. Three text-only messages with unique bodies.
2. Compute budget as rune length of the expected 2-of-3 block.

```go
import (
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("", "", "OLD_UNIQUE"),
		msg("", "", "MID_UNIQUE"),
		msg("", "", "NEW_UNIQUE"),
	}
	// Budget = exact size of the post-drop block (2 newest of 3).
	// Implementer must recompute header after drops: "showing 2 of 3".
	twoOfThree := "" +
		"Chat history (showing 2 of 3):\n" +
		"MID_UNIQUE\n" +
		"NEW_UNIQUE\n"
	req.Opts = msgfmt.Options{TotalBudgetRunes: utf8.RuneCountInString(twoOfThree)}
	return nil
}
```
