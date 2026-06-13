## Preconditions
- The mock config uses `llm_events` with the neutral `AgentEvent` format from `agent/event/types`.
- fake-codex converts each AgentEvent to its native codex JSON events.
- `llm_events` does NOT auto-detect native codex format (unlike `stdout_events`).
- `stdout_events` is deprecated; a stderr warning is emitted when it is present.

## Steps
1. Mark the test mode as llm-events.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=llm-events")
    return nil
}
```
