## Preconditions
- The mock config contains a `tool_call` AgentEvent with tool=bash.

## Steps
1. Write a mock config with a bash tool_call event that echoes a known string.
2. Run fake-opencode with JSON output.

```go
import (
    "encoding/json"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_real","llm_events":[{"type":"tool_call","tool":"bash","tool_input":{"command":"echo hello real bash"}}]}`)
    return nil
}
```
