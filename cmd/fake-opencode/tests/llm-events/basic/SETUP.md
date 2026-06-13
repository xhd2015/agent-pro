## Preconditions
- The mock config contains a `think` event in `llm_events`.

## Steps
1. Run fake opencode with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-opencode","session_id":"sess_basic","llm_events":[{"type":"think","text":"thinking about it"}]}`)
    return nil
}
```
