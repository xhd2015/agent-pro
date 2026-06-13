## Preconditions
- A test file exists on disk at a known path.
- The mock config contains a read tool_call AgentEvent.

## Steps
1. Create a test file and write a mock config with a read tool_call event.
2. Run fake-opencode and verify the event contains the file content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    content := "hello file content for read test\nline 2"
    filePath := createTestFile(t, req, "test-read-file.txt", content)
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_read_real","stdout_events":[{"type":"tool_call","tool":"read","tool_input":{"path":"` + filePath + `"}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
