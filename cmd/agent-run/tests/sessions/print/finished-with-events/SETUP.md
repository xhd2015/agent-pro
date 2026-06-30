# Scenario

**Feature**: print finished session with events — human-readable trace on stdout

```
seed meta.status=finished + events.jsonl -> sessions fake-codex/web_test123 --print -> trace lines
```

## Preconditions

- Session exists under `AGENT_RUN_HOME` with `status=finished`.
- `events.jsonl` has multiple `AgentEvent` lines including assistant message text.

## Steps

1. Seed session `fake-codex/web_test123` with status `finished`.
2. Append message events with known text `Hello from test` and a second line.
3. Run `agent-run sessions fake-codex/web_test123 --print`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	store := openAgentStore(t, req)
	seedSessionMeta(t, store, printRunner, printSessionID, "finished")
	appendAgentMessage(t, store, printRunner, printSessionID, "Hello from test")
	appendAgentMessage(t, store, printRunner, printSessionID, "Second trace line")
	req.Args = printSessionArgs(printRunner, printSessionID)
	return nil
}
```