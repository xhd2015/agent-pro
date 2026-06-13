## Preconditions
- The mock config contains legacy opencode events with `part` field.

## Steps
1. Run fake opencode with the legacy mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_old","stdout_events":[{"type":"text","part":{"id":"p1","type":"text","text":"legacy opencode ok"}}]}`)
    return nil
}
```
