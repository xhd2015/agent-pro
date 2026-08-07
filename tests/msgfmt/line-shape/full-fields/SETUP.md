# Scenario

**Feature**: id + sender + text produces the canonical line

```
Message{ID:m1, Sender:alice, Text:hello}
  -> Chat history (1 message):
     message_id=m1  [alice] : hello
```

## Preconditions

- Two spaces between `message_id=<id>` and `[sender]`.
- Space before and after `:` between sender bracket and body.

## Steps

1. One short message with all fields set.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{msg("m1", "alice", "hello")}
	req.Opts = msgfmt.Options{}
	return nil
}
```
