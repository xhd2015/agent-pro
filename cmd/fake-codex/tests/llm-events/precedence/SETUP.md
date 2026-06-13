## Preconditions
- The mock config contains both `llm_events` and `stdout_events` with different content.
- `llm_events` takes precedence.

## Steps
1. Run fake codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"think","text":"from llm"}],"stdout_events":[{"type":"think","text":"from stdout"}]}`)
    return nil
}
```
