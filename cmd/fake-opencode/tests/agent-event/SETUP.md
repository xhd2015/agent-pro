## Preconditions
- The mock config uses the neutral `AgentEvent` format for `stdout_events`.
- fake-opencode detects the format and converts each AgentEvent to its native opencode JSON events before emitting.
- Session ID from the mock config is injected into every emitted event.

## Steps
1. Mark the test mode as agent-event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    req.Env = append(req.Env, "FAKE_OPENCODE_TEST_MODE=agent-event")
    return nil
}
```
