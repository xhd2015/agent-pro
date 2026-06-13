## Preconditions
- The mock config contains a write tool_call AgentEvent with path and content.

## Steps
1. Write a mock config with a write tool_call event.
2. Run fake-opencode and verify the file was created on disk.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    targetPath := req.TempDir + "/write-output.txt"
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_write_real","stdout_events":[{"type":"tool_call","tool":"write","tool_input":{"path":"` + targetPath + `","content":"written content for verification"}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
