# Scenario

**Feature**: agent-run event-bus L2 harness — flags, NotifyTTYStarted,
AppendEventBusFlags, open-path publish policy

```
# help
agent-run run -h -> --event-bus-url + --event-bus-token

# notify (best-effort)
NotifyTTYStarted(opts, session, runner, workspace)
  -> type agent.tty.started / source agent-run / payload fields
  -> empty URL: no HTTP; failure: warning: only

# append flags
AppendEventBusFlags(args, url, token) -> argv parity for ForceNew follow-up

# open path
NotifyOnOpenPath(new-terminal) -> notify once + flags in follow-up
NotifyOnOpenPath(send)         -> no publish
```

## Preconditions

- Nested DOCTEST root under `cmd/agent-run/tests/event-bus/` (does not inherit
  parent `cmd/agent-run/tests` Setup/Run).
- Product APIs under design (RED until implementer):
  - `agentruncli.EventBusOpts` (`URL`, `Token`, `PublishHook`, `WarnWriter`;
    production may also hold `*eventbus.Publisher`)
  - `PublishHook func(ctx, eventType, source string, payload json.RawMessage) error`
    — L2 inject (tests); when nil, production uses eventbus.NewPublisher
  - `agentruncli.NotifyTTYStarted(opts, sessionID, runner, workspace)`
  - `agentruncli.NotifyOnOpenPath(kind, opts, sessionID, runner, workspace)` —
    `kind` is `"new-terminal"` (notify) or `"send"` (no-op)
  - `agentruncli.AppendEventBusFlags(args, url, token) []string`
  - `agentruncli.RunHelpText() string` — `run -h` body (must document event-bus flags)
  - `run` help / flag parse for `--event-bus-url` and `--event-bus-token`
- Shared client (product only): `github.com/xhd2015/dot-pkgs/go-pkgs/eventbus`.
  Module needs `replace` to brought `go-pkgs` tree. Doctest harness under
  `cmd/` may not inherit that replace for testcase imports — tests use wire
  string literals + PublishHook instead of importing eventbus.
- Parallel-safe: no `os.Setenv` / `Chdir`; no process stdio reassignment.
  Help uses pure `agentruncli.RunHelpText()` (no Handle stdio).
  Each leaf owns its `HTTPCapture` / inject publisher.
- Payload wire shape for `agent.tty.started` (locked by these tests):

```json
{"session_id":"<id>","runner":"<runner>","workspace":"<dir>"}
```

## Steps

1. Root `Setup` seeds default session identity and empty Capture.
2. Branch/leaf `Setup` sets `Op`, URL/token, mock options, OpenKind.
3. Root `Run` dispatches help / notify / append-flags / open-path.
4. Leaf `Assert` checks Response fields and HTTP observations.

## Context

- Source constant: `eventbus.SourceAgentRun` (`"agent-run"`).
- Type constant: `eventbus.TypeAgentTTYStarted` (`"agent.tty.started"`).
- Warning prefix on publish failure: `warning:` (stderr / WarnWriter).
- Empty URL is the disabled publisher path (no HTTP).

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
		req.SessionID = "sess-event-bus-1"
	}
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.Workspace == "" {
		req.Workspace = "/tmp/workspace-event-bus"
	}
	return nil
}
```
