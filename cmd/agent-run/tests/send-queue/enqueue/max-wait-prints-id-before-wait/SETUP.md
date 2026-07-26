# Scenario

**Feature**: `--max-wait` prints id before blocking for delivery

```
busy terminal + --max-wait 10s -> id line on stdout within ~1s, then blocks
```

## Steps

1. Set `req.Action = "max-wait-prints-id-before-wait"`.
2. Set `req.SendMessage = "max-wait-probe"`.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "max-wait-prints-id-before-wait"
	req.SendMessage = "max-wait-probe"
	return nil
}
```