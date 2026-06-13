## Preconditions
- The mock config contains a grep tool_use event with a `"mock"` object.

## Steps
1. Create a real file that would have grep matches.
2. Write a mock config with `"mock":{"output":"fake grep result\nfake_file.txt:1: fake match"}`.
3. Run fake-opencode and verify mock output is used instead of real grep.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    // Create a file that would match if real grep ran
    createTestFile(t, req, "grep-mock/file.txt", "this has REAL_MATCH in it")
    searchDir := req.TempDir + "/grep-mock"
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_grep_mock","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"grep","callID":"call_1","state":{"status":"pending","title":"mock grep","input":{"pattern":"REAL_MATCH","path":"` + searchDir + `"}}},"mock":{"output":"fake grep result\nfake_file.txt:1: fake match","exit_code":0}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
