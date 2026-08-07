# Scenario

**Feature**: empty Sender omits `[…]` brackets cleanly

```
Message{ID:m9, Sender:"", Text:solo}
  -> message_id=m9 : solo
```

## Preconditions

- No empty `[]` brackets and no double spaces left by a missing sender.

## Steps

1. One message with id and text only.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{msg("m9", "", "solo")}
	req.Opts = msgfmt.Options{}
	return nil
}
```
