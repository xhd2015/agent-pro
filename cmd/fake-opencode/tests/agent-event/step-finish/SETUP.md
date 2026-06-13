## Preconditions
- The mock config contains a `step_finish` AgentEvent.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_sf","stdout_events":[{"type":"step_finish"}]}`)
	return nil
}
```
