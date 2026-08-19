# Scenario

**Feature**: finished boxed grok-tty chrome classifies idle + empty composer

```
detach + live │ ❯ … │ chrome
  -> tty status --json at T+2s
  -> screen_status=idle, input_box=empty, sendable=true
```

Crime scene had `screen=banner` and `input_box=occupied` on this chrome.

## Steps

1. Fixed session id; observe at settle (before timeout+grace would reap).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "classify"
	req.SessionID = "idle-exit-classify-1"
	req.ObserveAfter = classifySettle
	return nil
}
```
