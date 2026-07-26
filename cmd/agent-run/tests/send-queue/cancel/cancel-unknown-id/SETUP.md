# Scenario

**Feature**: cancel unknown message id fails

```
send cancel session msg_9999 (not in queue) -> exit 1, stderr not found
```

## Steps

1. Set `req.Action = "cancel-unknown-id"`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "cancel-unknown-id"
	return nil
}
```