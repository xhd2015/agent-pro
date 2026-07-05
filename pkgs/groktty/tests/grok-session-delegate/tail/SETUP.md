# Scenario

**Feature**: TailUpdatesFromOffset emits grok_session-rich AgentEvents

```
# poll temp updates.jsonl from offset 0
updates.jsonl -> TailUpdatesFromOffset -> emit(AgentEvent with tool_call_id, status, turn_index, ActionDone)
```

## Preconditions

- Leaf provides synthetic wire lines ending in `turn_completed` so tail exits promptly.
- Events must include grok_session fields after refactor (RED before).

## Steps

1. Set `req.Target = "tail"`.
2. Leaf SETUPs populate `req.WireLines` and `req.UpdatesPath`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Target = "tail"
	req.UpdatesPath = newTempUpdatesPath(t)
	return nil
}
```