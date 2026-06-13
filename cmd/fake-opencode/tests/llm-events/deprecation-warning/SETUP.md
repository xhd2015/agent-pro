## Preconditions
- The mock config uses `stdout_events` (no `llm_events`).
- The deprecated field should still work but emit a stderr warning.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_depr","stdout_events":[{"type":"think","text":"deprecated still works"}]}`)
    return nil
}
```
