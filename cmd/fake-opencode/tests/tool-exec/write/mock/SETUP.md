## Preconditions
- The mock config contains a write tool_use event with a `"mock"` object.

## Steps
1. Write a mock config with `"tool":"write"` and `"mock":{"output":"mocked write done"}`.
2. Run fake-opencode.
3. Verify mock output is used and the file was **not** actually created.

```go
import (
    "os"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    mockJSON := `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_write_mock","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"write","callID":"call_1","state":{"status":"pending","title":"mock write","input":{"path":"` + req.TempDir + `/should-not-exist.txt","content":"this should not be written"}}},"mock":{"output":"mocked write done","exit_code":0}}]}`
    writeMockConfig(t, req, mockJSON)
    return nil
}
```
