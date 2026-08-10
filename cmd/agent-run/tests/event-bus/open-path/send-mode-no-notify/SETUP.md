# Scenario

**Feature**: live send path never publishes `agent.tty.started`

```
OpenKind=send + URL set + inject publisher ready
  -> NotifyOnOpenPath("send", …) is a no-op
  -> PublishCount=0
```

## Steps

1. OpenKind send; URL still set (would publish if dispatch wrongly always-notifies).
2. Run calls NotifyOnOpenPath("send", …); must not publish.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.OpenKind = openKindSend
	req.BuildFollowUpArgs = false
	req.EventBusURL = "http://127.0.0.1:23891"
	req.EventBusToken = "tok-o2"
	req.SessionID = "sess-open-o2"
	req.UseInjectPublisher = true
	return nil
}
```
