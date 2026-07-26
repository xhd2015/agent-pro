# Scenario

**Feature**: scenario screen_status idle detected by tty status

```
screen_status:idle -> tty status screen_status idle
```

## Steps

1. Configure leaf-specific `req` fields.
2. `Run` executes the scenario.

```go
import (
	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "scenario-declared-screen-status"
	return nil
}
```
