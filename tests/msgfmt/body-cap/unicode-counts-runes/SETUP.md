# Scenario

**Feature**: cap counts Unicode runes, not bytes

```
MaxPerMessageRunes=3, Text="你好世界" (4 runes, multi-byte UTF-8)
  -> body "你好…"  (3 runes)
```

## Preconditions

- Each CJK character is one rune (3 bytes in UTF-8) but counts as 1 toward the cap.
- A byte-based truncate would mis-cut mid-string differently; we lock rune semantics.

## Steps

1. Set max=3 and a four-rune Chinese body.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{msg("m1", "alice", "你好世界")}
	req.Opts = msgfmt.Options{MaxPerMessageRunes: 3}
	return nil
}
```
