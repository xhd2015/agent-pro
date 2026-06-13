## Preconditions
- A test file exists with known content.
- The mock config contains a grep tool_call AgentEvent searching for a non-existent pattern.

## Steps
1. Create a file and write a mock config with grep searching for a non-existent pattern.
2. Run fake-opencode and verify empty/missing output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    content := "some random content\nnothing to match here"
    filePath := createTestFile(t, req, "grep-no-match/test.txt", content)
    _ = filePath
    searchDir := req.TempDir + "/grep-no-match"
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_grep_nf","stdout_events":[{"type":"tool_call","tool":"grep","tool_input":{"pattern":"ZZZZ_NO_MATCH_ZZZZ","path":"` + searchDir + `"}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
