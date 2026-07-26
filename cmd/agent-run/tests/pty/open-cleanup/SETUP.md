# Scenario

**Feature**: harness reclaim after `run --open` leaves no leftover `__serve`

```
agent-run run --agent-runner grok-tty --open "probe"
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  + AGENT_RUN_GROK_TTY_COMMAND=fake hold TUI
  -> keep-alive __serve registered under AGENT_RUN_HOME
  -> harness reclaimServesUnderHome / ReclaimSessionID
  -> serve PID gone (no PTY leak for this session/home)
```

## Preconditions

- Product KeepAlive semantics for real users are unchanged; this branch tests
  **test harness cleanup** that reclaims serves started under the leaf home.
- Instant attach hook required (no interactive controlling TTY in CI).

## Steps

1. Grouping enables open-cleanup mode defaults.
2. Leaf finalizes prompt / fake TUI hold.
3. `Run` executes open, records serve, reclaims, reports liveness.
4. Assert open succeeded enough to start a serve, and reclaim cleared it.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "open-cleanup"
	req.OpenInstantAttach = true
	if req.GrokTTYCommand == "" {
		req.GrokTTYCommand = fakeTUIHoldSeconds(30)
	}
	if req.Prompt == "" {
		req.Prompt = "pty-open-cleanup"
	}
	if req.ExecTimeout < 60*time.Second {
		req.ExecTimeout = 90 * time.Second
	}
	return nil
}
```
