# Scenario

**Feature**: `--exit-on-idle` reaps grok-tty after timeout on finished chrome

```
detach --exit-on-idle --idle-timeout=2s + live boxed chrome
  -> wait 2s + 5s grace + slack
  -> tty not live
```

Crime scene stayed up past 10s+ (watchdog never armed because classify failed).

## Steps

1. Fixed session id; observe after timeout+grace.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "exit"
	req.SessionID = "idle-exit-gone-1"
	req.ObserveAfter = exitObserve
	return nil
}
```
