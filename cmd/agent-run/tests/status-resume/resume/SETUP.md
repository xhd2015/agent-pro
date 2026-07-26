# Scenario

**Feature**: `agent-run resume` — gate checks then run shortcut with provider `--resume`

```
seed meta -> agent-run resume [flags] (<session-id> | --grok-session-id ID) ["followup"]
  denied: not exited (hint send, not already-in-use)
         | unbound | missing session
         | --no-submit without --open
  ready + dead terminal: headless followup / --open accepted / by-grok-session-id
  ready + zombie registry: reclaim terminal id then reserve (not already-in-use)
```

## Steps

1. Leaf seeds meta / live / zombie fixtures and sets `req.Args` for resume.
2. `Run` executes CLI; assert gate errors, reclaim success, or argv/session success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Leaves finalize seed + Args for resume invocations.
	if len(req.Args) == 0 {
		req.Args = []string{"resume"}
	}
	return nil
}
```
