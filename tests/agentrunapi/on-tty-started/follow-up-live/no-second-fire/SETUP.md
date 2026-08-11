# Scenario

**Feature**: second AutoSendOrResume while live does not fire OnTTYStarted again

```
ModeRun establish -> HookCount=1
ModeSend follow-up -> HookCount stays 1; SendCalls=1
```

## Steps

1. Unique session for the two-step scenario.
2. Terminal / runner session ids for live seed after first run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.SessionID = "sess-tty-no-second"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-tty-no-second"
	req.TerminalID = "term-tty-no-second"
	req.RunnerSessID = "runner-tty-no-second"
	req.Prompt = "first establish"
	return nil
}
```
