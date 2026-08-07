# Scenario

**Feature**: body of exactly 1000 runes is unchanged under default cap

```
Text = 1000×'a', Options{} 
  -> body still 1000 'a's, BodiesTruncated=0
  DefaultMaxPerMessageRunes == 1000
```

## Preconditions

- Boundary: length == default max is **not** truncated.
- Zero `MaxPerMessageRunes` selects the default constant.

## Steps

1. Build body with `runeRepeat("a", 1000)`.
2. Use zero Options.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	body := runeRepeat("a", 1000)
	req.Msgs = []msgfmt.Message{msg("m1", "alice", body)}
	req.Opts = msgfmt.Options{}
	return nil
}
```
