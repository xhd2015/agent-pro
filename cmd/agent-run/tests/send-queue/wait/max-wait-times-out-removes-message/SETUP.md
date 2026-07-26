# Scenario

**Feature**: `--max-wait` timeout removes only the caller message

```
permanently busy + --max-wait 2s -> exit 1, stderr timeout, queue lacks msg id
```

## Steps

1. Set `req.Action = "max-wait-times-out-removes-message"`.
2. Set `req.SendMessage = "timeout-probe"`.
3. Set `req.ExecTimeout = 10s`.

```go
import (
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "max-wait-times-out-removes-message"
	req.SendMessage = "timeout-probe"
	req.ExecTimeout = 10 * time.Second
	return nil
}
```