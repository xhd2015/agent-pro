## Preconditions
- The mock config contains a read tool_call AgentEvent with a `mock` object.

## Steps
1. Write a mock config with mock content.
2. Run fake-opencode and verify mock content is used.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    createTestFile(t, req, "should-not-read.txt", "this should not appear")
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_read_mock","stdout_events":[{"type":"tool_call","tool":"read","mock":{"content":"fake read content","exit_code":0}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
