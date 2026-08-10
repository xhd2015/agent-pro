# Scenario

**Feature**: successful new-terminal open notifies once and puts flags in follow-up argv

```
OpenKind=new-terminal + URL set
  -> NotifyTTYStarted once (PublishCount=1)
  -> AppendEventBusFlags on follow-up base args
```

## Steps

1. OpenKind new-terminal; BuildFollowUpArgs true.
2. Token set so follow-up includes both flags.
3. Session identity for payload.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.OpenKind = openKindNewTerminal
	req.BuildFollowUpArgs = true
	req.EventBusURL = "http://127.0.0.1:23891"
	req.EventBusToken = "tok-o1"
	req.SessionID = "sess-open-o1"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-open-o1"
	req.BaseArgs = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", "sess-open-o1",
		"--agent-runner", "grok-tty",
	}
	req.UseInjectPublisher = true
	return nil
}
```
