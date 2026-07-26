# Scenario

**Feature**: send to unknown session fails without stdout id

```
agent-run send bogus-id "hi" -> exit 1, stderr not found, no msg_N stdout
```

## Steps

1. Set `req.Action = "session-not-found"`.
2. Set `req.SendMessage = "hi"`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "session-not-found"
	req.SendMessage = "hi"
	return nil
}
```