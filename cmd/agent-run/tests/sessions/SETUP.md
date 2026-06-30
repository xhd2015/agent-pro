# Scenario

**Subcommand**: `sessions` — list stored sessions or print one session's events

```
agent-run sessions [--json] -> list under AGENT_RUN_HOME/sessions
agent-run sessions <runner>/<id> --print -> read meta + events.jsonl -> FormatState trace
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- `AGENT_RUN_HOME` is isolated per test.
- Print leaves seed data via `pkgs/agentstorage` without a full `run` subprocess.

## Steps

1. Grouping `list/` or `print/` `Setup` sets `req.Args` for that operation mode.
2. Leaf `Setup` seeds sessions or finalizes flags.
3. `Run` executes `agent-run` with accumulated `req.Args`.
4. `Assert` checks exit code, stdout/stderr shape.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

const (
	printRunner    = "fake-codex"
	printSessionID = "web_test123"
)

func openAgentStore(t *testing.T, req *Request) agentstorage.Store {
	t.Helper()
	store, err := agentstorage.NewFileStore(req.Home)
	if err != nil {
		t.Fatalf("NewFileStore(%q): %v", req.Home, err)
	}
	return store
}

func seedSessionMeta(t *testing.T, store agentstorage.Store, runner, sessionID, status string) {
	t.Helper()
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    status,
	}
	if err := store.CreateSession(runner, sessionID, meta); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
}

func appendAgentMessage(t *testing.T, store agentstorage.Store, runner, sessionID, text string) {
	t.Helper()
	ev := types.AgentEvent{
		Type: types.ActionMessage,
		Role: "assistant",
		Text: text,
	}
	if err := store.AppendEvent(runner, sessionID, ev); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
}

func Setup(t *testing.T, req *Request) error {
	if req.SessionRunner == "" {
		req.SessionRunner = printRunner
	}
	return nil
}
```