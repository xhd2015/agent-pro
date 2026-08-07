# Scenario

**Feature**: body longer than 1000 runes is hard-capped with `…`

```
Text = 1001×'a', Options{}
  -> body = 999×'a' + "…"  (1000 runes total)
  BodiesTruncated=1
```

## Preconditions

- Effective max is `DefaultMaxPerMessageRunes` when Options field is 0.
- Final body rune count is exactly 1000 including the marker.

## Steps

1. Build body with 1001 ASCII `a` runes.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	body := runeRepeat("a", 1001)
	req.Msgs = []msgfmt.Message{msg("m1", "alice", body)}
	req.Opts = msgfmt.Options{}
	return nil
}
```
