## Preconditions
- The mock config contains a bash tool_use event.
- No `"mock"` object present, so real execution happens.

## Steps
1. Write a mock config with a bash tool_use event that echoes a known string.
2. Run fake-opencode with JSON output.
3. Verify the emitted event contains the real command output.

```go
import (
    "encoding/json"
    "testing"
)

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_real","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"bash","callID":"call_1","state":{"status":"pending","title":"echo test","input":{"command":"echo hello real bash"}}}}]}`)
    return nil
}
```
