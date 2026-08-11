# Scenario

**Feature**: empty event-bus URL + library wire → no HTTP

```
EventBusURL="" + WireOnTTYStarted + ModeRun
  -> PublishCount=0; no warning
```

## Steps

1. Empty URL; still install inject publisher so any accidental publish is recorded.
2. ModeRun via library-wire Op.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = ""
	req.EventBusToken = ""
	req.SessionID = "sess-wire-w2"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-wire-w2"
	req.UseInjectPublisher = true
	return nil
}
```
