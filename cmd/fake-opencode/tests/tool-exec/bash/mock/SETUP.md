## Preconditions
- The mock config contains a bash tool_use event **with** a `"mock"` object.
- The `"mock"` object provides fake output/exit_code.

## Steps
1. Write a mock config with a bash event and `"mock":{"output":"fake mock output","exit_code":0}`.
2. The event's `input.command` would be something that would produce different output if executed (e.g., `echo real output`).
3. Run fake-opencode and verify the mock output is used instead of real output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_mock","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"bash","callID":"call_1","state":{"status":"pending","title":"mock bypass","input":{"command":"echo REAL OUTPUT SHOULD NOT APPEAR"}}},"mock":{"output":"fake mock output","exit_code":0,"stderr":""}}]}`)
    return nil
}
```
