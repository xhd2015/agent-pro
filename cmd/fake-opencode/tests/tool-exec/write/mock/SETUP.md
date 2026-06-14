## Preconditions
- The mock config contains a write tool_call AgentEvent with a `mock` object.

## Steps
1. Write a mock config with mock output.
2. Run fake-opencode and verify the file was not created.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_write_mock","llm_events":[{"type":"tool_call","tool":"write","mock":{"output":"mocked write done"}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
