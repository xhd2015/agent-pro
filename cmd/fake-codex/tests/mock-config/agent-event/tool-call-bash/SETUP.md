## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=bash and a command.
- No mock output is set, so the command executes for real.

## Steps
1. Create a file to verify real execution.
2. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","stdout_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo hello-from-bash"}}]}`)
    return nil
}
```
