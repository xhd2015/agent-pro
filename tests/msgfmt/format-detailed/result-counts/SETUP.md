# Scenario

**Feature**: Result count fields after MaxMessages selection

```
5 msgs, MaxMessages=2
  -> SourceCount=5, Shown=2, OldestDropped=3
```

## Preconditions

- Counts describe selection outcome independent of label wording (label still asserted).

## Steps

1. Five short messages; MaxMessages=2.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msgs := make([]msgfmt.Message, 0, 5)
	for i := 1; i <= 5; i++ {
		msgs = append(msgs, msg(fmt.Sprintf("id%d", i), "s", fmt.Sprintf("t%d", i)))
	}
	req.Msgs = msgs
	req.Opts = msgfmt.Options{MaxMessages: 2}
	return nil
}
```
