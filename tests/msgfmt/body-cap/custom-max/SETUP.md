# Scenario

**Feature**: explicit `MaxPerMessageRunes` overrides the default

```
MaxPerMessageRunes=5, Text="abcdefgh"
  -> body "abcd…"  (4 + marker = 5 runes)
```

## Preconditions

- Custom max applies even when far below 1000.
- Marker still costs one rune inside the budget.

## Steps

1. Set `Opts.MaxPerMessageRunes = 5`.
2. Body is eight ASCII letters.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{msg("m1", "alice", "abcdefgh")}
	req.Opts = msgfmt.Options{MaxPerMessageRunes: 5}
	return nil
}
```
