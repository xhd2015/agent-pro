# Scenario

**Feature**: `--detach` keep-alive uses the serve idle watchdog

```
agent-run run --detach --exit-on-idle --idle-timeout=2s
  --agent-runner=grok-tty --agent-runner-binary llm-mock-run-grok
  -> idle-policy.json + live boxed chrome
```

## Steps

1. Wire detach + 2s idle + live chrome hook (holds PTY 30s unless watchdog exits).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	configureLiveChromeHook(req, 30)
	req.IdleTimeout = defaultIdleTimeout
	return nil
}
```
