## Preconditions
- The mock config contains a `"tool":"write"` tool_use event with path and content in the input.

## Steps
1. Write a mock config with a write tool_use event.
2. Run fake-opencode.
3. After execution, verify the file was actually created on disk with the specified content.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    targetPath := req.TempDir + "/write-output.txt"
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_write_real","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"write","callID":"call_1","state":{"status":"pending","title":"write file","input":{"path":"` + targetPath + `","content":"written content for verification"}}}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
