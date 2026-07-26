# Scenario

**Feature**: cancel removes pending message before delivery

```
--no-wait enqueue on busy session -> send cancel msg_N -> exit 0, never injected
```

## Steps

1. Set `req.Action = "cancel-pending-message"`.
2. Set `req.SendMessage = "cancel-probe"`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "cancel-pending-message"
	req.SendMessage = "cancel-probe"
	return nil
}
```