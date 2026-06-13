## Preconditions
- The mock config uses only `llm_events` (no `stdout_events`).

## Steps
1. Run fake codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"think","text":"clean codex"}]}`)
    return nil
}
```
