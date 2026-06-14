## Preconditions
- The mock config contains a single `think` AgentEvent with text.

## Steps
1. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"think","text":"analyzing the request"}]}`)
    return nil
}
```
