# Scenario

**Feature**: URL set + true-TTY ModeRun via WireOnTTYStarted publishes once

```
EventBusURL set + WireOnTTYStarted + AutoSendOrResume ModeRun
  -> PublishCount=1
  -> type=agent.tty.started source=agent-run
  -> payload session_id / runner / workspace
```

## Steps

1. Non-empty EventBusURL; inject publisher.
2. Distinct session id for payload assert.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = "http://127.0.0.1:23892"
	req.EventBusToken = ""
	req.SessionID = "sess-wire-w1"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-wire-w1"
	req.UseInjectPublisher = true
	return nil
}
```
