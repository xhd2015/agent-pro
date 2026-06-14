## Preconditions
- The mock config contains a grep tool_call AgentEvent with a `mock` object.

## Steps
1. Create a real file that would have grep matches.
2. Write a mock config with mock output.
3. Run fake-opencode and verify mock output is used instead of real grep.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    _ = createTestFile(t, req, "grep-mock/file.txt", "this has REAL_MATCH in it")
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_grep_mock","llm_events":[{"type":"tool_call","tool":"grep","mock":{"output":"fake grep result\nfake_file.txt:1: fake match","exit_code":0}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
