## Preconditions
- The mock config contains a bash tool_use event that runs `exit 3`.

## Steps
1. Run fake-opencode with JSON output.
2. Verify the event has non-zero exit_code.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bash_fail","stdout_events":[{"type":"tool_use","part":{"id":"t1","type":"tool","tool":"bash","callID":"call_1","state":{"status":"pending","title":"failing command","input":{"command":"exit 3"}}}}]}`)
    return nil
}
```
