## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=bash and a `mock` field providing fake output.
- The mock overrides the exit code and output.

## Steps
1. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo ignored"},"mock":{"output":"mocked-output","exit_code":2}}]}`)
    return nil
}
```
