## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=bash and a `mock` field providing fake output.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_mock","stdout_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo ignored"},"mock":{"output":"mocked-stdout","stderr":"mocked-stderr","exit_code":1}}]}`)
    return nil
}
```
