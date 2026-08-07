# Scenario

**Feature**: `LastMessageID` is the newest input message id

```
msgs m-old, m-mid, m-new; MaxMessages=1 (only m-new shown)
  -> LastMessageID == "m-new"
```

## Preconditions

- LastMessageID comes from the **input** newest message, which is also the
  trigger kept under MaxMessages.

## Steps

1. Three messages; MaxMessages=1 so only the last is shown.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{
		msg("m-old", "a", "old"),
		msg("m-mid", "b", "mid"),
		msg("m-new", "c", "new"),
	}
	req.Opts = msgfmt.Options{MaxMessages: 1}
	return nil
}
```
