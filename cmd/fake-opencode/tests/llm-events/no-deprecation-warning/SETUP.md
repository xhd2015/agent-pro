## Preconditions
- The mock config uses only `llm_events` (no `stdout_events`).
- No deprecation warning should appear.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_clean","llm_events":[{"type":"think","text":"clean"}]}`)
    return nil
}
```
