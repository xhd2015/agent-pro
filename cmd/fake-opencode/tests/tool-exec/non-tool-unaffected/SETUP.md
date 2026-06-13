## Preconditions
- The mock config contains only message and error AgentEvents (no tool_call).

## Steps
1. Write a mock config with one message event and one error event.
2. Run fake-opencode and verify these events pass through converted to opencode format.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_nontool","stdout_events":[{"type":"message","text":"plain text message"},{"type":"error","text":"an error occurred"}]}`)
    return nil
}
```
