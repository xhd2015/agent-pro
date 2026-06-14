## Preconditions
- The mock config contains an `error` AgentEvent with text and a nonzero exit code.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_err","exit_code":4,"stderr":"planned failure","llm_events":[{"type":"error","text":"operation failed"}]}`)
    return nil
}
```
