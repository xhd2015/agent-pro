# Scenario

**Feature**: ForceNew + library OnTTYStarted share once-guard → exactly one publish

```
AlreadyNotified shared
  -> NotifyOnOpenPath(new-terminal)
  -> WireOnTTYStarted callback
  -> PublishCount=1 (not 2)
```

## Steps

1. URL + token set; inject publisher.
2. Session identity for payload.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = "http://127.0.0.1:23893"
	req.EventBusToken = "tok-d1"
	req.SessionID = "sess-double-d1"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-double-d1"
	req.UseInjectPublisher = true
	req.UseAlreadyNotified = true
	return nil
}
```
