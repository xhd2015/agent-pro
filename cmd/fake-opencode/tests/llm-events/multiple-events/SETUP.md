## Preconditions
- The mock config contains three events in `llm_events`: think, message, done.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_multi","llm_events":[{"type":"think","text":"thinking"},{"type":"message","text":"hello"},{"type":"done"}]}`)
    return nil
}
```
