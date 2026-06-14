## Preconditions
- The mock config contains two message events in a specific order.

## Steps
1. Run fake Codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"message","text":"first response"},{"type":"message","text":"second response"}]}`)
    return nil
}
```

