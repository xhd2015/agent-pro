# Scenario

**Feature**: `--open` with empty prompt succeeds when TUI never paints ready markers

```
agent-run run --agent-runner grok-tty --open
  + fake TUI: "booting" only, sleep 12, exit
  + no AGENT_RUN_OPEN_ATTACH_INSTANT
  -> exit 0
  -> no "banner not detected" / "TUI banner not detected"
  -> stderr once "grok-tty: <id>" after attach returns
```

## Preconditions

- Empty prompt: no inject expected.
- No INSTANT: exercises production open readiness (attach-first), not the
  INSTANT soft-wait branch that already ignored banner timeouts.

## Steps

1. Inherit open grouping no-banner hold + `--open` with no positional prompt.
2. Assert success without banner error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	clearOpenInstantAttach(req)
	req.Prompt = ""
	// Explicit empty open args (no trailing prompt).
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open"}
	hold := writeFakeTUINoBannerHold(t, req.TempDir, 12)
	setGrokTTYCommand(req, hold)
	return nil
}
```
