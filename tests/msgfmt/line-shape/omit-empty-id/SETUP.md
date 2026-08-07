# Scenario

**Feature**: empty ID omits `message_id=` prefix cleanly

```
Message{ID:"", Sender:bob, Text:ping}
  -> [bob] : ping
```

## Preconditions

- No empty `message_id=` token and no dangling spaces from a missing id.

## Steps

1. One message with empty ID, non-empty sender and text.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{msg("", "bob", "ping")}
	req.Opts = msgfmt.Options{}
	return nil
}
```
