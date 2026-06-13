## Preconditions
- A test file exists on disk at a known path.
- The mock config contains a `"tool":"read"` tool_use event pointing to that file.

## Steps
1. Create a test file with known content.
2. Write a mock config with a read tool_use event referencing that file.
3. Run fake-opencode and verify the event contains the file content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    content := "hello file content for read test\nline 2"
    filePath := createTestFile(t, req, "test-read-file.txt", content)
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_read_real","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"read","callID":"call_1","state":{"status":"pending","title":"read file","input":{"path":"` + filePath + `"}}}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
