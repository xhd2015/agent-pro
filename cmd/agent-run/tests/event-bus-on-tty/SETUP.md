# Scenario

**Feature**: agent-run wires OnTTYStarted to event-bus HTTP; no double-fire

```
# library wire
WireOnTTYStarted(opts) + AutoSendOrResume(ModeRun)
  -> publish when URL set; empty URL -> no HTTP

# double-fire
NotifyOnOpenPath(new-terminal) + WireOnTTYStarted(info)
  + shared AlreadyNotified
  -> PublishCount == 1
```

## Preconditions

- Nested DOCTEST root under `cmd/agent-run/tests/event-bus-on-tty/` (does not
  inherit sibling `event-bus` Setup/Run).
- Product APIs under design (RED until implementer):
  - `agentruncli.EventBusOpts.AlreadyNotified *bool`
  - `agentruncli.WireOnTTYStarted(opts) func(agentrunapi.TTYStartedInfo)`
  - `agentrunapi.TTYStartedInfo` + `Opts.OnTTYStarted`
  - Existing: `NotifyTTYStarted`, `NotifyOnOpenPath`, PublishHook inject
- L2 only: PublishHook capture; temp store; no real network / iTerm / agent binary.
- Parallel-safe: no `t.Setenv` / `os.Chdir`; each leaf owns Capture + temp Home.
- agent-pro stays neutral to ai-critic.
- Payload wire shape:

```json
{"session_id":"<id>","runner":"<runner>","workspace":"<dir>"}
```

## Steps

1. Root `Setup` seeds default session identity and empty Capture.
2. Branch `Setup` sets Op (library-wire | double-fire).
3. Leaf `Setup` sets URL / token / identity.
4. Root `Run` exercises wire or double-fire; leaf Assert checks publish count.

## Context

- Type: `agent.tty.started`; source: `agent-run`.
- Empty URL disables publish (no HTTP, no warning).
- Sibling regression tree: `cmd/agent-run/tests/event-bus/` (must stay GREEN).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	if req.SessionID == "" {
		req.SessionID = "sess-event-bus-on-tty-1"
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.Workspace == "" {
		req.Workspace = "/tmp/workspace-event-bus-on-tty"
	}
	return nil
}
```
