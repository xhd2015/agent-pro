# Scenario

**Feature**: print finished session with events — human-readable trace on stdout

```
seed meta.status=finished + events.jsonl -> sessions web_test123 --print -> trace lines
```

## Preconditions

- Session exists under flat `AGENT_RUN_HOME/sessions/<id>/` with `status=finished`.
- `events.jsonl` has multiple `AgentEvent` lines including assistant message text.

## Steps

1. Seed session `web_test123` (runner meta `fake-codex`) with status `finished`.
2. Append message events with known text `Hello from test` and a second line.
3. Run `agent-run sessions web_test123 --print`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printSessionID, "finished")
	appendAgentMessage(t, store, printSessionID, "Hello from test")
	appendAgentMessage(t, store, printSessionID, "Second trace line")
	req.Args = printSessionArgs(printSessionID)
	return nil
}
```
