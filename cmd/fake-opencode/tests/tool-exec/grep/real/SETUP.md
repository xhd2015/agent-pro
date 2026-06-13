## Preconditions
- A test file exists on disk containing a known search pattern.
- The mock config contains a `"tool":"grep"` tool_use event.

## Steps
1. Create a file with content including `UNIQUE_MARKER_FOR_GREP`.
2. Write a mock config with a grep tool_use event searching for that pattern.
3. Run fake-opencode and verify real grep output appears in the event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    content := "line 1: no match\nline 2: UNIQUE_MARKER_FOR_GREP is here\nline 3: no match either"
    filePath := createTestFile(t, req, "grep-test/file.txt", content)
    searchDir := req.TempDir + "/grep-test"
    _ = filePath
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_grep_real","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"grep","callID":"call_1","state":{"status":"pending","title":"grep search","input":{"pattern":"UNIQUE_MARKER_FOR_GREP","path":"` + searchDir + `"}}}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
