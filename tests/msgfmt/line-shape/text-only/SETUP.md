# Scenario

**Feature**: empty ID and Sender leave only the body on the line

```
Message{ID:"", Sender:"", Text:just text}
  -> just text
```

## Preconditions

- No `message_id=`, no brackets, no leading `: `.

## Steps

1. One message with text only.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Msgs = []msgfmt.Message{msg("", "", "just text")}
	req.Opts = msgfmt.Options{}
	return nil
}
```
