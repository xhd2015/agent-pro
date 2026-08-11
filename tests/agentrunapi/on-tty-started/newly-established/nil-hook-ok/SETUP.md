# Scenario

**Feature**: nil OnTTYStarted does not panic; ModeRun continues

```
OnTTYStarted=nil + ModeRun
  -> RunSession once; no panic; HookCount=0
```

## Steps

1. Leave InstallHook false (nil OnTTYStarted).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.InstallHook = false
	req.SessionID = "sess-tty-nil-hook"
	req.Prompt = "run without hook"
	return nil
}
```
