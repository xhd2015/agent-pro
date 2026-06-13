## Preconditions
- The mock config contains a bash tool_use event with a `"mock"` object that sets a custom exit_code and stderr.

## Steps
1. Write a mock config with `"mock":{"output":"","exit_code":42,"stderr":"custom error message"}`.
2. Run fake-opencode and verify exit_code 42 and stderr appear in the event.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_mock_exit","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"bash","callID":"call_1","state":{"status":"pending","title":"mock exit code","input":{"command":"exit 0"}}},"mock":{"output":"","exit_code":42,"stderr":"custom error message"}}]}`)
    return nil
}
```
