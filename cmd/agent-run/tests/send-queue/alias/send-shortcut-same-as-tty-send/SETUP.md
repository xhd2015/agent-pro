# Scenario

**Feature**: send shortcut and tty send share queue id semantics

```
send -> msg_1; tty send -> msg_2 on same session
```

## Steps

1. Set `req.SendMessage = "alias-probe"`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SendMessage = "alias-probe"
	return nil
}
```