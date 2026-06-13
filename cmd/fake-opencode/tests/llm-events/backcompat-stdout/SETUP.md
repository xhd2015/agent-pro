## Preconditions
- The mock config uses `stdout_events` with agent event format (no `llm_events`).
- Backward compatibility: `stdout_events` must still parse and convert events.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_bcompat","stdout_events":[{"type":"think","text":"backward compat works"}]}`)
    return nil
}
```
