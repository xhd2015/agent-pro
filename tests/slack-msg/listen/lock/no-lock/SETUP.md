# Scenario

**Feature**: `--no-lock` disables singleton

```
slack-msg listen --no-lock -> banner lock (none)
  -> second --no-lock does not report "already running"
```

## Steps

1. Pass `--no-lock` via harness `NoLock`.
2. Start daemon + second instance.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.NoLock = true
	req.Daemon = true
	req.SecondInstance = true
	req.InjectEvents = nil
	req.WantAgentCalls = 0
	return nil
}
```
