# Scenario

**Feature**: MaxMessages above source count shows every message

```
3 messages, MaxMessages=5
  -> showing 3 of 3; all three bodies present
```

## Preconditions

- Cap is a maximum, not a pad: never invent messages.

## Steps

1. Three short fully-fielded messages; MaxMessages=5.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("m1", "a", "one"),
		msg("m2", "b", "two"),
		msg("m3", "c", "three"),
	}
	req.Opts = msgfmt.Options{MaxMessages: 5}
	return nil
}
```
