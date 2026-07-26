# Scenario

**Feature**: `sessions --print --grok-session-id UUID` prints events for the
matched grok-tty session (same trace as bare-id path)

```
seed finished grok-tty + runner_session_id + events
  -> agent-run sessions --print --grok-session-id UUID
  -> exit 0; formatted trace includes message text
```

## Preconditions

- Flat `sessions/<id>/meta.json` with `runner=grok-tty` and non-empty
  `runner_session_id`.
- `events.jsonl` has known assistant message text.

## Steps

1. Seed session `print_gsid_s1` (runner `grok-tty`) with status `finished` and
   provider UUID.
2. Append message events with known text.
3. Run `agent-run sessions --print --grok-session-id <UUID>` (no positional id).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const (
	printGrokSessionID     = "print_gsid_s1"
	printGrokRunnerUUID    = "550e8400-e29b-41d4-a716-446655440920"
	printGrokMessageText   = "Hello from grok-session-id print"
	printGrokSecondMessage = "Second gsid print line"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	store := openAgentStore(t, req)
	req.SessionID = printGrokSessionID
	req.SessionRunner = "grok-tty"
	seedSessionMetaRunnerSessionID(t, store, "grok-tty", printGrokSessionID, "finished", printGrokRunnerUUID)
	appendAgentMessage(t, store, printGrokSessionID, printGrokMessageText)
	appendAgentMessage(t, store, printGrokSessionID, printGrokSecondMessage)
	req.Args = printGrokSessionArgs(printGrokRunnerUUID)
	return nil
}
```
