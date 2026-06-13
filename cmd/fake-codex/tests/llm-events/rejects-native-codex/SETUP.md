## Preconditions
- The mock config has `llm_events` containing a native codex format event (with `item` field).
- `llm_events` only accepts neutral AgentEvent format — native codex format must not be auto-detected.

## Steps
1. Run fake codex with the mock config.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
    writeMockConfig(t, req, `{"version":"agent-pro.fake-runner.v1","runner":"fake-codex","llm_events":[{"type":"item.started","item":{"id":"1","type":"reasoning","text":"native format"}}]}`)
    return nil
}
```
