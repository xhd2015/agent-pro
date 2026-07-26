# Scenario

**Feature**: MODE=run when session is missing (or unbound / non-resume-ready)

```
auto + new --session-id + prompt
  -> normal agentui.Run path; creates session; no provider --resume
```

## Steps

1. Ensure argv-sensitive leaves clear the grok-tty command hook env.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Run-mode creates sessions via provider; never inherit ambient TTY command hook.
	req.GrokTTYCommand = ""
	req.Env = withoutEnvKey(req.Env, envGrokTTYCommand)
	return nil
}
```
