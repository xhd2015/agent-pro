# Scenario

**Feature**: `--detach` keep-alive uses the serve idle watchdog

```
agent-run run --detach --exit-on-idle --idle-timeout=10s
  --agent-runner=codex-tty --agent-runner-binary llm-mock-run-codex
  -> idle-policy.json + live Codex TUI
```

## Steps

1. Wire detach + 10s idle (watchdog samples at 0, 5s, 10s).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.IdleTimeout = defaultIdleTimeout
	req.ObserveAfter = defaultIdleTimeout + defaultGrace + probeSlack
	return nil
}
```
