# Scenario

**Feature**: `run --open` after attach returns — finalize grok session discovery/bind

```
agent-run run --agent-runner grok-tty --open [prompt]
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  + optional GROK_HOME / AGENT_RUN_GROK_TTY_GROK_SESSION_ID seed
  -> after attach: print grok session lines + persist runner_session_id
     OR error not resolved (exit ≠ 0)
```

## Preconditions

- Instant attach hook required for CI (no interactive TTY).
- Prefer fake TUI via `AGENT_RUN_GROK_TTY_COMMAND`.
- Discovery success uses temp `GROK_HOME` + known UUID hook.

## Steps

1. Leaf configures open flags, discovery fixtures, instant attach.
2. `Run` executes open path; optionally re-reads meta.
3. Assert stderr session lines or not-resolved error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.OpenInstantAttach = true
	req.Runner = "grok-tty"
	if req.GrokTTYCommand == "" {
		req.GrokTTYCommand = fakeTUIRespondHi()
	}
	return nil
}
```
