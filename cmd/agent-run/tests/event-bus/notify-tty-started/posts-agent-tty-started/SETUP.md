# Scenario

**Feature**: NotifyTTYStarted posts `agent.tty.started` with source `agent-run` and session payload

```
NotifyTTYStarted(URL set, token empty)
  -> one Publish; type=agent.tty.started; source=agent-run
  -> payload.session_id / runner / workspace
```

## Steps

1. Set non-empty EventBusURL (token empty).
2. Use inject publisher recording body.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EventBusURL = "http://127.0.0.1:23891"
	req.EventBusToken = ""
	req.SessionID = "sess-notify-n2"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-notify-n2"
	req.UseInjectPublisher = true
	return nil
}
```
