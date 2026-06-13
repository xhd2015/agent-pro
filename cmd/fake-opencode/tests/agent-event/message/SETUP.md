## Preconditions
- The mock config contains a `message` AgentEvent with text.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_msg","stdout_events":[{"type":"message","text":"all tasks complete"}]}`)
    return nil
}
```
