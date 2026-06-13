## Preconditions
- The mock config contains a read tool_use event with a `"mock"` object.
- The mock provides fake `content` without accessing the filesystem.

## Steps
1. Write a mock config with `"tool":"read"` and `"mock":{"content":"fake read content"}`.
2. The input.path points to a valid file, but mock should bypass it.
3. Run fake-opencode and verify mock content is used.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    // Create a real file that should NOT be read (mock bypasses it)
    createTestFile(t, req, "should-not-read.txt", "this should not appear")
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_read_mock","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"read","callID":"call_1","state":{"status":"pending","title":"mock read","input":{"path":"` + req.TempDir + `/should-not-read.txt"}}},"mock":{"content":"fake read content","exit_code":0}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
