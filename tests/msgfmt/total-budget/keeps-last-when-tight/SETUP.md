# Scenario

**Feature**: even a tiny budget still keeps the last (trigger) message

```
msgs: m1/FIRST_DROP_ME, m2/TRIGGER_LAST (id set, empty sender)
TotalBudgetRunes=1  (smaller than any non-empty block)
  -> still shows message_id=m2 : TRIGGER_LAST; drop first
  -> Shown=1; last body present
```

## Preconditions

- Prefer keeping the last message over satisfying the budget when conflict.
- Last message remains subject to per-message body cap (not relevant here).
- Two messages with id, no sender → line shape `message_id=<id> : <text>`.

## Steps

1. Two messages with id and empty sender; budget of 1 rune.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("m1", "", "FIRST_DROP_ME"),
		msg("m2", "", "TRIGGER_LAST"),
	}
	req.Opts = msgfmt.Options{TotalBudgetRunes: 1}
	return nil
}
```
