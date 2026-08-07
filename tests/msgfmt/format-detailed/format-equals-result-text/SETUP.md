# Scenario

**Feature**: `Format` always equals `FormatDetailed(...).Text`

```
non-trivial msgs + options
  Format(msgs, opts) == FormatDetailed(msgs, opts).Text
```

## Preconditions

- Parity holds under selection + truncation together.

## Steps

1. Mix MaxMessages and MaxPerMessageRunes on four messages.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("1", "a", "short"),
		msg("2", "b", "this body is definitely longer than eight"),
		msg("3", "c", "mid"),
		msg("4", "d", "tail-message"),
	}
	req.Opts = msgfmt.Options{
		MaxMessages:        3,
		MaxPerMessageRunes: 8,
	}
	return nil
}
```
