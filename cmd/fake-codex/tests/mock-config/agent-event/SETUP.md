## Preconditions
- The mock config uses the neutral `AgentEvent` format for `stdout_events`.
- fake-codex detects the format and converts each AgentEvent to its native codex events before emitting.

## Steps
1. Mark the test mode as agent-event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_CODEX_TEST_MODE=agent-event")
    return nil
}
```
