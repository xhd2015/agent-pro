## Preconditions
- The mock config uses `stdout_events` with the neutral AgentEvent format.
- A `sleep` event type with `delay_ms` produces no output and pauses event emission.
- Individual events may carry a `delay_ms` field for a pre-emission delay.

## Steps
1. Inherit the fake-opencode binary and environment from the root SETUP.md.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    t.Log("setting up delay test group")
    return nil
}
```
