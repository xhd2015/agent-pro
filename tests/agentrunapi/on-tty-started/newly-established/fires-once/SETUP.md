# Scenario

**Feature**: first ModeRun with OnTTYStarted set fires the hook once

```
OnTTYStarted set + ModeRun (new session)
  -> HookCount=1
  -> info.SessionID == opts.SessionID
```

## Steps

1. Install recording OnTTYStarted.
2. Unique session id for payload assert.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InstallHook = true
	req.SessionID = "sess-tty-fire-once"
	req.Runner = "grok-tty"
	req.Workspace = "/tmp/ws-tty-fire-once"
	req.Prompt = "open first TTY"
	return nil
}
```
