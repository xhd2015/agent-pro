## Preconditions
- The mock config contains one text event.

## Steps
1. Run fake opencode with JSON output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_text","stdout_events":[{"type":"text","part":{"id":"p1","type":"text","text":"fake opencode answered"}}]}`)
    return nil
}
```

