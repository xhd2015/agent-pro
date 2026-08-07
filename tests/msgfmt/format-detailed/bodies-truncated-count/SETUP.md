# Scenario

**Feature**: `BodiesTruncated` counts how many bodies were shortened

```
3 msgs, MaxPerMessageRunes=4
  bodies: "ab" (ok), "abcdefgh" (trunc), "xyzw" (exactly 4, ok)
  -> BodiesTruncated=1
```

## Preconditions

- Exact-length bodies do not count as truncated.
- Only over-max bodies increment the counter.

## Steps

1. Three messages with the bodies above; max=4.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("a", "s", "ab"),
		msg("b", "s", "abcdefgh"),
		msg("c", "s", "xyzw"),
	}
	req.Opts = msgfmt.Options{MaxPerMessageRunes: 4}
	return nil
}
```
