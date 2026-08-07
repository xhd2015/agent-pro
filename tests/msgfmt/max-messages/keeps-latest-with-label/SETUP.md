# Scenario

**Feature**: MaxMessages keeps the latest three of ten and labels K of N

```
10 messages (m01..m10), MaxMessages=3
  -> only m08,m09,m10
  -> header Chat history (showing 3 of 10):
```

## Preconditions

- Oldest seven are dropped (`OldestDropped=7`).
- Order among shown lines remains chronological (oldest-of-shown first).

## Steps

1. Build ten messages with distinct bodies `b01`…`b10`.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msgs := make([]msgfmt.Message, 0, 10)
	for i := 1; i <= 10; i++ {
		msgs = append(msgs, msg(
			fmt.Sprintf("m%02d", i),
			"u",
			fmt.Sprintf("b%02d", i),
		))
	}
	req.Msgs = msgs
	req.Opts = msgfmt.Options{MaxMessages: 3}
	return nil
}
```
